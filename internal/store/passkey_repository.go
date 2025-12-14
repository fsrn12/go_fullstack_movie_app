package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"multipass/internal/model"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasskeyStore interface {
	GetUserByEmail(ctx context.Context, email string) (*model.PasskeyUser, error)
	GetUserByID(ctx context.Context, ID int) (*model.PasskeyUser, error)
	SaveUser(ctx context.Context, user model.PasskeyUser) error

	GenSessionID() (string, error)
	GetSession(token string) (webauthn.SessionData, bool)
	SaveSession(token string, data webauthn.SessionData)
	DeleteSession(token string)
}

type PasskeyRepository struct {
	db       *pgxpool.Pool
	logger   logging.Logger
	sessions map[string]webauthn.SessionData
}

// NewPasskeyRepository initializes a new PasskeyRepository with a database connection.and logger
func NewPasskeyRepository(db *pgxpool.Pool, logger logging.Logger) *PasskeyRepository {
	return &PasskeyRepository{
		db:       db,
		logger:   logger,
		sessions: make(map[string]webauthn.SessionData),
	}
}

// GetSession retrieves session data from the in-memory map.
func (r *PasskeyRepository) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// GetSession retrieves session data from the in-memory map.
func (r *PasskeyRepository) GetSession(token string) (webauthn.SessionData, bool) {
	r.logger.Info(fmt.Sprintf("GetSession: %v", r.sessions[token]))
	val, ok := r.sessions[token]
	return val, ok
}

// SaveSession stores session data in the in-memory map.
func (r *PasskeyRepository) SaveSession(token string, data webauthn.SessionData) {
	r.logger.Info(fmt.Sprintf("SaveSession: %s - %v", token, data))
	r.sessions[token] = data
}

// DeleteSession removes session data from the in-memory map.
func (r *PasskeyRepository) DeleteSession(token string) {
	r.logger.Info("DeleteSession: %v", token)
	delete(r.sessions, token)
}

// GenSessionID generates a random session ID.

// GetUserByEmail retrieves user by email.
func (r *PasskeyRepository) GetUserByEmail(ctx context.Context, email string) (*model.PasskeyUser, error) {
	op := "PasskeyRepository.GetUserByEmail"
	meta := common.Envelop{
		"email": email,
		"op":    op,
	}
	r.logger.Info(fmt.Sprintf("GetUser: %v", email))

	// STEP 1: CHECK IF USER EXISTS BY EMAIL
	var userID int
	err := r.db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		return nil, handleDatabaseError(err, r.logger, op, "user", meta)
	}

	// STEP 2: FETCH USER CREDENTIALS FROM PASSKEYS TABLE
	rows, err := r.db.Query(ctx, "SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		meta["user_id"] = userID
		return nil, handleDatabaseError(err, r.logger, op, "user_id", meta)
	}

	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.logger.Error("Failed to scan passkey row", err)
			return nil, err
		}

		cred, err := deserializeCredential(keys)
		if err != nil {
			r.logger.Error("failed to deserialize credential", err)
			continue // SKIPPING INVALID CREDENTIAL
		}

		credentials = append(credentials, cred)
	}

	// STEP 3: CONSTRUCT AND RETURN PASSKEYUSER
	user := model.PasskeyUser{
		ID:           []byte(strconv.Itoa(userID)), // CONVERT int ID to byte slice
		Name:         email,
		DisplayName:  email,
		Credientials: credentials,
	}

	return &user, nil
}

// GetUserByEmail retrieves user by ID.
func (r *PasskeyRepository) GetUserByID(ctx context.Context, ID int) (*model.PasskeyUser, error) {
	op := "PasskeyRepository.GetUserByID"
	meta := common.Envelop{
		"op": op,
	}
	r.logger.Info(fmt.Sprintf("GetUser: %v", ID))

	// STEP 1: CHECK IF USER EXISTS BY ID
	var userID int
	var email string
	err := r.db.QueryRow(ctx, "SELECT id, email FROM users WHERE user_id = $1", ID).Scan(&userID, &email)
	if err != nil {
		return nil, handleDatabaseError(err, r.logger, op, "user", meta)
	}

	meta["user_id"] = userID
	meta["email"] = email
	// STEP 2: FETCH USER CREDENTIALS FROM PASSKEYS TABLE
	rows, err := r.db.Query(ctx, "SELECT keys FROM passkeys WHERE user_id = $1", ID)
	if err != nil {
		return nil, handleDatabaseError(err, r.logger, op, "user_id", meta)
	}

	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.logger.Error("Failed to scan passkey row", err)
			return nil, err
		}

		cred, err := deserializeCredential(keys)
		if err != nil {
			r.logger.Error("failed to deserialize credential", err)
			continue // SKIPPING INVALID CREDENTIAL
		}

		credentials = append(credentials, cred)
	}

	// STEP 3: CONSTRUCT AND RETURN PASSKEYUSER
	user := model.PasskeyUser{
		ID:           []byte(strconv.Itoa(ID)), // CONVERT int ID to byte slice
		Name:         email,
		DisplayName:  email,
		Credientials: credentials,
	}

	return &user, nil
}

func (r *PasskeyRepository) SaveUser(ctx context.Context, user model.PasskeyUser) error {
	op := "PasskeyRepository.SaveUser"
	meta := common.Envelop{
		"op": op,
	}

	// STEP 1: CONVERT user ID from byte slice to integer
	userID, err := strconv.Atoi(string(user.ID))
	if err != nil {
		r.logger.Error("invalid user ID", err)
		return fmt.Errorf("invalid user ID: %+w", err)
	}

	// STEP 2: INSERT NEW CREDENTIALS
	for _, cred := range user.Credientials {
		keys, err := serializeCredential(cred)
		if err != nil {
			r.logger.Error("failed to serialize credentials", err)
			continue
		}

		// CHECK IF KEY ALREADY EXISTS IN DATABASE
		var exists bool
		err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM passkeys WHERE user_id = $1 AND keys = $2)", userID, keys).Scan(&exists)
		if err != nil {
			r.logger.Error("failed to check if passkey exit", err)
			continue
		}

		if exists {
			r.logger.Info(fmt.Sprintf("Passkey already exists for user_id: %d", userID))
			return apperror.ErrUniqueViolation(err, op, r.logger, meta)
		}

		// FINALLY INSERT KEY INTO DB
		_, err = r.db.Exec(ctx, "INSERT INTO passkeys (user_id, keys) VALUES ($1, $2)", userID, keys)
		if err != nil {
			return handleDatabaseError(err, r.logger, op, "passkey", meta)
		}
	}

	return nil
}

// serializeCredential converts a WebAuthn credential to a JSON string.
func serializeCredential(credential webauthn.Credential) (string, error) {
	data, err := json.Marshal(credential)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %+w", err)
	}

	return string(data), nil
}

// deserializeCredential converts a JSON string back to a WebAuthn credential.
func deserializeCredential(data string) (webauthn.Credential, error) {
	var credential webauthn.Credential
	err := json.Unmarshal([]byte(data), &credential)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("failed to unmarshal credential: %+w", err)
	}

	return credential, nil
}

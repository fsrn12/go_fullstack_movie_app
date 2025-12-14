package store

import (
	"context"
	"fmt"

	"multipass/internal/auth/tokens"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStore interface {
	SaveRefreshToken(ctx context.Context, token *tokens.Token) error
	GetRefreshTokenHash(ctx context.Context, hash []byte) ([]byte, error)
	GetTokenDetailsByHash(ctx context.Context, tokenHash []byte) (*tokens.Token, error)
	DeleteRefreshToken(ctx context.Context, userID int, token []byte) error
	DeleteAllTokensForUser(ctx context.Context, userID int) error

	// Save verification/reset token
	SaveAuthToken(ctx context.Context, token *tokens.Token) error

	// Retrieve verification/reset token
	GetAuthTokenByHash(ctx context.Context, tokenHash []byte, scope string) (*tokens.Token, error)

	// Delete a token by hash (used after successful consumption)
	DeleteAuthTokenByHash(ctx context.Context, tokenHash []byte) error
}

type TokenRepository struct {
	db     *pgxpool.Pool
	logger logging.Logger
}

func NewTokenStore(db *pgxpool.Pool, logger logging.Logger) TokenStore {
	return &TokenRepository{
		db:     db,
		logger: logger,
	}
}

// SaveRefreshToken stores a new hashed refresh token in the database.
func (t *TokenRepository) SaveRefreshToken(ctx context.Context, refreshToken *tokens.Token) error {
	op := getOp(QuerySaveRefreshToken)
	meta := common.Envelop{
		"context": op,
	}

	query, err := getQuery(QuerySaveRefreshToken, t.logger, meta)
	if err != nil || query == "" {
		return err
	}
	_, err = t.db.Exec(ctx, query,
		refreshToken.Hash,
		refreshToken.UserID,
		refreshToken.Expiry,
		refreshToken.Scope,
	)
	if err != nil {
		meta["user_id"] = refreshToken.UserID
		return handleDatabaseError(err, t.logger, op, "user", meta)
	}

	return nil
}

func (t *TokenRepository) GetRefreshTokenHash(ctx context.Context, hash []byte) ([]byte, error) {
	op := getOp(QueryGetRefreshTokenHash)
	meta := common.Envelop{
		"context": op,
	}

	query, err := getQuery(QueryGetRefreshTokenHash, t.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}

	var dbHash []byte
	err = t.db.QueryRow(ctx, query, hash).Scan(&dbHash)
	if err != nil {
		return nil, handleDatabaseError(err, t.logger, op, "dbRefreshTokenHash", meta)
	}

	return dbHash, nil
	// if err != nil {
	// 	tokenErr := apperror.ErrTokenMissing(err, t.logger, nil)
	// 	if err == sql.ErrNoRows {
	// 		t.logger.Error("refresh token not found", tokenErr)
	// 		return nil, tokenErr
	// 	}

	// 	t.logger.Error("failed to query refresh token hash", err)
	// 	return nil, fmt.Errorf("failed to query for refresh token hash: %+w", err)
	// }

	// t.logger.Info("Successfully found token hash")
}

func (t *TokenRepository) DeleteRefreshToken(ctx context.Context, userID int, hashedToken []byte) error {
	op := getOp(QueryDeleteRefreshToken)
	meta := common.Envelop{
		"context": op,
		"user_id": userID,
	}

	query, err := getQuery(QueryDeleteRefreshToken, t.logger, meta)
	if err != nil || query == "" {
		return err
	}

	_, err = t.db.Exec(ctx, query, userID, hashedToken)
	if err != nil {
		return handleDatabaseError(err, t.logger, op, "delete dbRefreshTokenHash", meta)
	}

	return nil
}

func (t *TokenRepository) DeleteAllTokensForUser(ctx context.Context, userID int) error {
	op := getOp(QueryDeleteAllTokensForUser)
	meta := common.Envelop{
		"context": op,
		"user_id": userID,
	}
	query, err := getQuery(QueryDeleteAllTokensForUser, t.logger, meta)
	if err != nil || query == "" {
		return err
	}
	_, err = t.db.Exec(ctx, query, userID)
	if err != nil {
		return handleDatabaseError(err, t.logger, op, "delete AllDBRefreshTokenHash", meta)
	}

	return nil
}

// GetTokenDetailsByHash retrieves a refresh token's details from the database by its hash.
func (t *TokenRepository) GetTokenDetailsByHash(ctx context.Context, tokenHash []byte) (*tokens.Token, error) {
	const op = "tokenStore.GetTokenDetailsByHash"
	query := `
		SELECT user_id, token_hash, expiry, scope
		FROM refresh_tokens
		WHERE token_hash = $1`

	token := &tokens.Token{}
	var storedHash []byte // To scan the BYTEA from DB

	row := t.db.QueryRow(ctx, query, tokenHash)
	err := row.Scan(&token.UserID, &storedHash, &token.Expiry, &token.Scope)
	if err != nil {
		return nil, handleDatabaseError(err, t.logger, "token_lookup", op, common.Envelop{
			"op":         op,
			"token_hash": fmt.Sprintf("%x", tokenHash),
		})
	}

	// Assign retrieved hash to the token struct
	token.Hash = storedHash

	return token, nil
}

func (t *TokenRepository) SaveAuthToken(ctx context.Context, token *tokens.Token) error {
	op := "TokenRepository.SaveAuthToken"
	meta := common.Envelop{
		"context": op,
	}

	query := `INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	_, err := t.db.Exec(ctx, query,
		token.Hash,
		token.UserID,
		token.Expiry,
		token.Scope,
	)
	if err != nil {
		meta["user_id"] = token.UserID
		return handleDatabaseError(err, t.logger, op, "user", meta)
	}

	return nil
}

// Retrieve verification/reset token
func (t *TokenRepository) GetAuthTokenByHash(ctx context.Context, tokenHash []byte, scope string) (*tokens.Token, error) {
	op := "TokenRepository.GetRefreshTokenHash"
	meta := common.Envelop{
		"context": op,
	}

	query := `SELECT user_id, hash, expiry, scope
        FROM tokens
        WHERE hash = $1 AND scope = $2
        LIMIT 1;`

	dbToken := &tokens.Token{}
	err := t.db.QueryRow(ctx, query, tokenHash, scope).Scan(&dbToken.UserID, &dbToken.Hash, &dbToken.Expiry, &dbToken.Scope)
	if err != nil {
		meta["error_detail"] = "token_not_found_or_db_error"
		return nil, handleDatabaseError(err, t.logger, op, "authTokenHash", meta)
	}

	return dbToken, nil
}

func (t *TokenRepository) DeleteAuthTokenByHash(ctx context.Context, tokenHash []byte) error {
	op := "TokenRepository.DeleteAuthTokenByHash"
	meta := common.Envelop{
		"context": op,
	}

	query := `DELETE FROM tokens WHERE hash = $1`

	_, err := t.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return handleDatabaseError(err, t.logger, op, "delete authTokenHash", meta)
	}

	return nil
}

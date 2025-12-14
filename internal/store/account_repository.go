package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"multipass/internal/model"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/go-webauthn/webauthn/webauthn"
)

/* AccountStore Interface */
type AccountStore interface {
	// Authenticate(ctx context.Context, email, password string) (bool, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (int, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByID(ctx context.Context, userID int) (*model.User, error)
	GetAccountDetails(ctx context.Context, email string) (*model.User, error)
	SaveCollection(ctx context.Context, userID int, movieID int, collection string) (bool, error)
	RemoveMovieFromCollection(ctx context.Context, userID int, movieID int, collection string) (bool, error)
	UpdateUser(ctx context.Context, user *model.User) error
	GetMovieList(ctx context.Context, list string, userID int) ([]model.Movie, error)
	SaveProfilePictureUrl(ctx context.Context, userID int, profilePictureUrl string) error
	MarkUserAsVerified(ctx context.Context, userID int) error
	UpdatePassword(ctx context.Context, userID int, passwordHash string) error
}

type AccountRepository struct {
	db     *pgxpool.Pool
	logger logging.Logger
}

func NewAccountRepository(db *pgxpool.Pool, logger logging.Logger) *AccountRepository {
	return &AccountRepository{
		db:     db,
		logger: logger,
	}
}

func (r *AccountRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (int, error) {
	// Create new user in the db
	var userID int
	op := "store.CreateUser"
	meta := common.Envelop{
		"user_email": email,
		"context":    op,
	}

	// Get the SQL query
	query, err := getQuery(QueryCreateUser, r.logger, meta)
	if err != nil || query == "" {
		return 0, err
	}

	// Execute query
	err = r.db.QueryRow(ctx, query, name, email, string(passwordHash), time.Now()).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique violation (duplicate email)
				r.logger.Errorf(fmt.Sprintf("failed to create user: email already in use: %s", email), nil, meta)
				return 0, apperror.ErrUniqueViolation(err, op, r.logger, meta)
			case "23506": // Foreign key violation
				r.logger.Errorf(fmt.Sprintf("failed to create user: foreign key violation: %s", email), nil, meta)
				return 0, apperror.ErrForeignKeyViolation(err, r.logger, meta)
			default:
				// Handle other Postgres errors here
				r.logger.Errorf("failed to create user: unexpected error", err, meta)
				return 0, apperror.ErrDatabaseOpFailed(apperror.CodeDatabaseError, apperror.ErrQueryFailedMsg, op, err, r.logger, meta)
			}
		}
	}

	// Return successfully created user ID
	return userID, nil
}

// FindUserByEmail retrieves a user by their email
func (r *AccountRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.findUser(ctx, QueryGetUserByEmail, email)
}

func (r *AccountRepository) FindUserByID(ctx context.Context, userID int) (*model.User, error) {
	return r.findUser(ctx, QueryGetUserByID, userID)
}

func (r *AccountRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	op := "store.CheckEmailExists"
	meta := common.Envelop{
		"email":         email,
		"context":       op,
		"db_error_type": "exists_check",
	}
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	if err != nil {
		return false, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}
	return exists, nil
}

func (r *AccountRepository) GetAccountDetails(ctx context.Context, email string) (*model.User, error) {
	meta := common.Envelop{
		"email": email,
		"op":    "store.GetAccountDetails",
	}

	user, err := r.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	favoriteCollection, err := r.getMovieList(ctx, QueryGetFavorite, user.ID)
	if err != nil {
		return nil, err
	}

	if len(favoriteCollection) == 0 {
		r.logger.Info("User has no item in favorite list", meta)
	}
	user.Favorites = favoriteCollection

	watchlistCollection, err := r.getMovieList(ctx, QueryGetWatchlist, user.ID)
	if err != nil {
		return nil, err
	}
	if len(watchlistCollection) == 0 {
		r.logger.Info("User has no items in watchlist", meta)
	}
	user.Watchlist = watchlistCollection

	return user, nil
}

func (r *AccountRepository) UpdateUser(ctx context.Context, user *model.User) error {
	op := "store.UpdateUser"
	meta := common.Envelop{
		"email":   user.Email,
		"user_id": user.ID,
		"context": op,
	}

	query, err := getQuery(QueryUpdateUserDetails, r.logger, meta)
	if err != nil || query == "" {
		return err
	}

	err = r.db.QueryRow(ctx, query, user.Name, user.Email, user.ProfilePictureURL).Scan(&user.UpdatedAt)
	if err != nil {
		return handleDatabaseError(err, r.logger, op, "user_update", meta)
	}

	return nil
}

func (r *AccountRepository) SaveProfilePictureUrl(ctx context.Context, userID int, profilePictureUrl string) error {
	op := "AccountRepository.SaveProfilePictureUrl"
	query := `UPDATE users SET profile_picture_url = $1 WHERE id = $2`
	metaData := common.Envelop{
		"user_id":             userID,
		"profile_picture_url": profilePictureUrl,
	}

	_, err := r.db.Exec(ctx, query, profilePictureUrl, userID)
	if err != nil {
		return handleDatabaseError(err, r.logger, op, "profilePictureUrl", metaData)
	}

	return nil
}

func (r *AccountRepository) SaveCollection(ctx context.Context, userID int, movieID int, collection string) (bool, error) {
	op := "store.SaveCollection"
	meta := common.Envelop{
		"user_id":    userID,
		"movie_id":   movieID,
		"collection": collection,
		"context":    op,
	}
	queryExists, err := getQuery(QueryIfMovieExists, r.logger, meta)
	if err != nil || queryExists == "" {
		return false, err
	}

	// Verify if relationship between userID and movieID already exists
	var exists bool
	err = r.db.QueryRow(ctx, queryExists, userID, movieID, collection).Scan(&exists)
	if err != nil {
		meta["db_error_type"] = "exists_check"
		r.logger.Error("failed to check if movie already exists", err, meta)
		return false, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}

	if exists {
		r.logger.Error("Movie already in the collection; no action taken", nil, meta)
		return true, nil // Returning true since the movie is already in the collection
	}

	// Add movie to collection
	queryAdd, err := getQuery(QueryAddToCollection, r.logger, meta)
	if err != nil || queryAdd == "" {
		return false, err
	}

	_, err = r.db.Exec(ctx, queryAdd, userID, movieID, collection, time.Now())
	if err != nil {
		meta["db_error_type"] = "insert_collection"
		return false, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}

	r.logger.Info("movie successfully added movie to user collection", meta)
	return true, nil
}

func (r *AccountRepository) RemoveMovieFromCollection(ctx context.Context, userID int, movieID int, collection string) (bool, error) {
	op := "store.RemoveMovieFromCollection"
	meta := common.Envelop{
		"user_id":    userID,
		"movie_id":   movieID,
		"collection": collection,
		"context":    op,
	}

	queryExists, err := getQuery(QueryIfMovieExists, r.logger, meta)
	if err != nil || queryExists == "" {
		return false, err
	}

	// Verify if relationship between userID and movieID already exists
	var exists bool
	err = r.db.QueryRow(ctx, queryExists, userID, movieID, collection).Scan(&exists)
	if err != nil {
		meta["db_error_type"] = "exists_check"
		r.logger.Error("failed to check if movie exists in collection", err, meta)
		return false, handleDatabaseError(err, r.logger, op, "movie_exists_query", meta)
	}

	if !exists {
		r.logger.Error("Movie does not exist in the collection; no action taken", nil, meta)
		return true, nil // Returning true since the movie is already absent from the collection
	}

	// Add movie to collection
	queryRemove, err := getQuery(QueryRemoveFromToCollection, r.logger, meta)
	if err != nil || queryRemove == "" {
		return false, err
	}

	_, err = r.db.Exec(ctx, queryRemove, userID, movieID, collection)
	if err != nil {
		meta["db_error_type"] = "delete_from_collection"
		r.logger.Error("failed to remove movie from collection", err, "meta", meta)
		return false, handleDatabaseError(err, r.logger, op, "movie_remove_from_collection", meta)
	}

	r.logger.Info("movie successfully removed movie from user collection", "meta", meta)
	return true, nil
}

func (r *AccountRepository) GetUserIDByEmail(ctx context.Context, email string) (*int, error) {
	user, err := r.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &user.ID, nil
}

// UpdateLastLogin Helper function to update the last login time
func (r *AccountRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	op := "store.UpdateLastLogin"
	meta := common.Envelop{
		"user_id": userID,
		"context": op,
	}

	query, err := getQuery(QueryUpdateLastLogin, r.logger, meta)
	if err != nil || query == "" {
		return err
	}

	_, err = r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		meta["db_error_type"] = "update_last_login"
		return apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}
	r.logger.Info("Successfully updated last login", meta)
	return nil
}

func (r *AccountRepository) GetMovieList(ctx context.Context, list string, userID int) ([]model.Movie, error) {
	op := "store.GetMovieList"
	meta := common.Envelop{
		"user_id": userID,
		"context": op,
	}
	var query string

	switch list {
	case "favorites":
		query = QueryGetFavorite
	case "watchlist":
		query = QueryGetWatchlist
	default:
		return nil, fmt.Errorf("invalid list name %s: it must be either 'favorites' or 'watchlist'", list)
	}

	query, err := getQuery(query, r.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return []model.Movie{}, err
	}
	defer rows.Close()

	// getting total number or rows
	numberOfRows, err := getRowCount(ctx, r.db, r.logger, "user_movies", "user_id", "favorites", userID)
	if err != nil {
		r.logger.Error("getRowCount failed to get number of rows", err)
		numberOfRows = 0
	}

	// scan Rows
	movieList, err := scanRowsToSlice(
		rows,
		scanMovie,
		r.logger,
		op,
		meta,
		numberOfRows)
	if err != nil {
		return nil, err
	}

	return movieList, nil
}

func (r *AccountRepository) getMovieList(ctx context.Context, queryKey string, userID int) ([]model.Movie, error) {
	op := getOp(queryKey)
	meta := common.Envelop{
		"user_id": userID,
		"context": op,
	}

	query, err := getQuery(queryKey, r.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return []model.Movie{}, err
	}
	defer rows.Close()

	// getting total number or rows
	numberOfRows, err := getRowCount(ctx, r.db, r.logger, "user_movies", "user_id", "favorites", userID)
	if err != nil {
		r.logger.Error("getRowCount failed to get number of rows", err)
		numberOfRows = 0
	}

	// scan Rows
	movieList, err := scanRowsToSlice(
		rows,
		scanMovie,
		r.logger,
		op,
		meta,
		numberOfRows)
	if err != nil {
		return nil, err
	}

	return movieList, nil
}

func (r *AccountRepository) findUser(ctx context.Context, queryKey string, arg any) (*model.User, error) {
	user := &model.User{}
	op := getOp(queryKey)
	meta := common.Envelop{
		"op": op,
	}

	if email, ok := arg.(string); ok {
		meta["email"] = email
	} else if userID, ok := arg.(int); ok {
		meta["user_id"] = userID
	}

	query, err := getQuery(queryKey, r.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}

	err = r.db.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHashed,
		&user.ProfilePictureURL,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.ConfirmedAt,
	)
	if err != nil {
		return nil, handleDatabaseError(err, r.logger, "user", op, meta)
	}

	return user, nil
}

func (r *AccountRepository) GetUserFromTokenHash(ctx context.Context, tokenHash, scope string) (*model.User, error) {
	user := &model.User{}
	op := "store.GetUserFromTokenHash"

	query, err := getQuery(QueryGetUserFromTokenHash, r.logger, nil)
	if err != nil || query == "" {
		return nil, err
	}

	err = r.db.QueryRow(ctx, query, tokenHash, scope, time.Now()).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.ConfirmedAt)
	if err != nil {
		return nil, handleDatabaseError(err, r.logger, op, "user", common.Envelop{
			"op":         op,
			"token_hash": tokenHash,
			"scope":      scope,
		})
	}

	return user, nil
}

func (r *AccountRepository) MarkUserAsVerified(ctx context.Context, userID int) error {
	op := "store.MarkUserAsVerified"
	meta := common.Envelop{"user_id": userID, "context": op}

	query := `
        UPDATE users
        SET is_verified = TRUE, time_confirmed = NOW(), updated_at = NOW()
        WHERE id = $1
    `

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return handleDatabaseError(err, r.logger, op, "mark_verified", meta)
	}

	r.logger.Info("User successfully marked as verified", meta)
	return nil
}

func (r *AccountRepository) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	op := "store.UpdatePassword"
	meta := common.Envelop{"user_id": userID, "context": op}

	query := `
        UPDATE users
        SET password_hashed = $1, updated_at = NOW()
        WHERE id = $2
    `

	_, err := r.db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return handleDatabaseError(err, r.logger, op, "update_password", meta)
	}
	r.logger.Info("User password successfully updated", meta)
	return nil
}

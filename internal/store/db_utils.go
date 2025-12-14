package store

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"multipass/internal/model"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Constants for valid tables and keys
const (
	TableUsers      = "users"
	TableMovies     = "movies"
	TableUserMovies = "user_movies"
	KeyUserID       = "user_id"
	KeyMovieID      = "movie_id"
	RelFavorites    = "favorites"
	RelWatchlist    = "watchlist"
)

func getOp(q string) string {
	return fmt.Sprintf("store.%s", strings.TrimPrefix(q, "Query"))
}

// Function to build dynamic query conditions
func buildQueryCondition(table, key, rel string, args []any) (string, []interface{}, error) {
	var queryBuilder strings.Builder
	var queryArgs []interface{}

	// Validate table and key
	validTable := map[string]string{
		TableUsers:      TableUsers,
		TableMovies:     TableMovies,
		TableUserMovies: TableUserMovies,
	}
	validKey := map[string]string{
		KeyUserID:  KeyUserID,
		KeyMovieID: KeyMovieID,
	}

	if _, ok := validTable[table]; !ok {
		return "", nil, fmt.Errorf("invalid table: %s", table)
	}
	if _, ok := validKey[key]; !ok {
		return "", nil, fmt.Errorf("invalid key: %s", key)
	}

	// Build the base query
	queryBuilder.WriteString("SELECT COUNT(*) FROM ")
	queryBuilder.WriteString(table)
	queryBuilder.WriteString(" WHERE ")

	// Add the primary condition
	queryBuilder.WriteString(table + "." + key + " = $1")
	queryArgs = append(queryArgs, args[0])

	// If there's a relation (favorites, watchlist), add it
	if rel != "" {
		queryBuilder.WriteString(" AND relation_type = $2")
		queryArgs = append(queryArgs, rel)
	}

	return queryBuilder.String(), queryArgs, nil
}

// scanRowsToSlice is a generic helper to scan pgx.Rows into a slice of type T.
// It takes:
//   - rows: The pgx.Rows object returned by a database query.
//   - scanFn: A function that knows how to scan a single row into an instance of type *T.
//   - log: Your application logger.
//   - op: The operation name for logging/error context.
//   - meta: Additional metadata for logging/error context.
//   - limit: An optional hint for initial slice capacity (can be 0 if unknown).

func scanRowsToSlice[T any](rows pgx.Rows, scanFn func(pgx.Rows, *T) error, log logging.Logger, op string, meta common.Envelop, limit int) ([]T, error) {
	results := make([]T, 0, limit)

	// Iterate over the rows and scan them into the results slice
	for rows.Next() {
		var item T
		if err := scanFn(rows, &item); err != nil {
			meta["db_error_type"] = "query_scan"
			log.Errorf("failed to scan row: %v", err)
			return nil, apperror.ErrDatabaseOpFailed(
				apperror.CodeDataException,
				apperror.ErrQueryFailedMsg,
				op,
				err,
				log,
				meta,
			)
		}

		results = append(results, item)
	}

	// Check for any error that might have occurred during the iteration
	if err := rows.Err(); err != nil {
		meta["db_error_type"] = "query_iteration"
		log.Errorf("Error during rows iteration: %v", err)
		return nil, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			log,
			meta,
		)
	}

	// Return the successfully scanned results
	return results, nil
}

func getRowCount(ctx context.Context, db *pgxpool.Pool, log logging.Logger, table string, key string, rel string, args ...any) (int, error) {
	op := "store.GetRowCount"
	meta := common.Envelop{
		"context":        op,
		"querying_table": table,
	}

	// Build the query dynamically based on the table, key, and optional relation type
	query, queryArgs, err := buildQueryCondition(table, key, rel, args)
	if err != nil {
		return 0, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			"invalid query parameters",
			op,
			err,
			log,
			meta,
		)
	}

	// Perform the database query
	var count int
	err = db.QueryRow(ctx, query, queryArgs...).Scan(&count)
	if err != nil {
		log.Errorf("failed to execute count query: %v", err)
		return 0, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			"failed to execute count query",
			op,
			err,
			log,
			meta,
		)
	}

	return count, nil
}

func scanUser(rows pgx.Rows, u *model.User) error {
	return rows.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHashed,
	)
}

// Scan function for Movie model
func scanMovie(rows pgx.Rows, m *model.Movie) error {
	return rows.Scan(
		&m.ID,
		&m.TMDB_ID,
		&m.Title,
		&m.Tagline,
		&m.ReleaseYear,
		&m.Overview,
		&m.Score,
		&m.Popularity,
		&m.Language,
		&m.PosterURL,
		&m.TrailerURL,
	)
}

// scanActor function (UPDATED: now scans 6 fields including Character)
func scanActor(rows pgx.Rows, a *model.Actor) error {
	return rows.Scan(
		&a.ID,
		&a.FirstName,
		&a.LastName,
		&a.ImageURL,
	)
}

func handleDatabaseError(err error, log logging.Logger, op string, key string, meta common.Envelop) error {
	metaData := common.Envelop{
		"op": op,
	}
	maps.Copy(meta, metaData)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Error(fmt.Sprintf("%s not found", key), err, metaData)
		return apperror.ErrRecordNotFound(err, log, metaData)
	}

	// Handle specific PostgreSQL error codes
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation (e.g., duplicate email)
			log.Error("unique constraint violation", err, metaData)
			return apperror.ErrUniqueViolation(err, op, log, metaData)

		case "23506": // Foreign key violation (e.g., invalid foreign reference)
			log.Error("foreign key violation", err, metaData)
			return apperror.ErrForeignKeyViolation(err, log, metaData)

		// Add more error codes as needed
		default:
			// Handle other PostgreSQL error codes
			log.Error("postgres error", err, metaData)
			return apperror.ErrDatabaseOpFailed(
				apperror.CodeDatabaseError,
				apperror.ErrQueryFailedMsg,
				op,
				err,
				log,
				metaData)
		}
	}

	// Handle other errors
	metaData["db_error_type"] = "query_scan"
	log.Error("failed to scan user: %+w", err, metaData)
	return apperror.ErrDatabaseOpFailed(
		apperror.CodeDatabaseError,
		apperror.ErrQueryFailedMsg,
		op,
		err,
		log,
		metaData)
}

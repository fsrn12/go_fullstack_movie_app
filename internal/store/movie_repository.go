package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"multipass/internal/model"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

/* MovieStore Interface */
type MovieStore interface {
	GetTopMovies(ctx context.Context) ([]model.Movie, error)
	GetRandomMovies(ctx context.Context) ([]model.Movie, error)
	GetMovieByID(ctx context.Context, id int) (model.Movie, error)
	SearchMovieByName(ctx context.Context, name string, order string, genre *int) ([]model.Movie, error)
	GetAllGenres(ctx context.Context) ([]model.Genre, error)
	DoesMovieExist(ctx context.Context, tmdbID int) (bool, error)
	SearchMovies(ctx context.Context, query string) ([]int, error)
	// SaveMovie(ctx context.Context, movie *model.Movie) error
}

type MovieRepository struct {
	db     *pgxpool.Pool
	logger logging.Logger
}

func NewMovieRepository(db *pgxpool.Pool, logger logging.Logger) *MovieRepository {
	return &MovieRepository{
		db:     db,
		logger: logger,
	}
}

// Close closes the database connection pool.
func (r *MovieRepository) Close() {
	if r.db != nil {
		r.db.Close()
		r.logger.Info("Database connection pool closed.")
	}
}

// DoesMovieExist checks if a movie with the given TMDB ID already exists in the database.
func (r *MovieRepository) DoesMovieExist(ctx context.Context, tmdbID int) (bool, error) {
	// This function uses a direct query string. If you want to use the map,
	// you'd use getQuery(QueryIfMovieExists, r.logger, nil)
	query := `SELECT EXISTS(SELECT 1 FROM movies WHERE tmdb_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, tmdbID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if movie exists: %w", err)
	}
	return exists, nil
}

// SearchMovies performs a full-text search on movies based on a query string.
// It searches across title, tagline, overview, and keywords.
// Returns a slice of TMDB IDs of matching movies, ordered by relevance.
func (r *MovieRepository) SearchMovies(ctx context.Context, query string) ([]int, error) {
	searchQuery := `SELECT tmdb_id FROM movies WHERE search_vector @@ plainto_tsquery('english', $1) ORDER BY ts_rank(search_vector, plainto_tsquery('english', $1)) DESC LIMIT 50;`

	rows, err := r.db.Query(ctx, searchQuery, query)
	if err != nil {
		return nil, fmt.Errorf("failed to perform movie search: %w", err)
	}
	defer rows.Close()

	var tmdbIDs []int
	for rows.Next() {
		var tmdbID int
		if err := rows.Scan(&tmdbID); err != nil {
			r.logger.Errorf("Failed to scan TMDB ID during search results: %v", err)
			continue
		}
		tmdbIDs = append(tmdbIDs, tmdbID)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating search results: %w", rows.Err())
	}

	return tmdbIDs, nil
}

// GetTopMovies retrieves top 10 movies
func (r *MovieRepository) GetTopMovies(ctx context.Context) ([]model.Movie, error) {
	return r.getMovies(ctx, QueryGetTopMovies, 10)
}

// GetRandomMovies retrieves 10 random movies
func (r *MovieRepository) GetRandomMovies(ctx context.Context) ([]model.Movie, error) {
	return r.getMovies(ctx, QueryGetRandomMovies, 10)
}

// GetMovieByID retrieves a movie by its ID
func (r *MovieRepository) GetMovieByID(ctx context.Context, id int) (model.Movie, error) {
	var m model.Movie
	op := getOp(QueryGetMovieByID)
	meta := common.Envelop{
		"movie_id": id,
		"context":  op,
	}

	row, err := r.queryRowWithErrorHandling(ctx, QueryGetMovieByID, meta, id)
	if err != nil {
		return model.Movie{}, err
	}

	err = row.Scan(
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
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Movie{}, apperror.ErrMovieNotFound(err, r.logger, meta)
		}

		return model.Movie{}, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			common.Envelop{"movie_id": id, "db_error_type": "query_scan"},
		)
	}

	// Fetching Movie Relations
	if err := r.FetchMovieRelations(ctx, &m); err != nil {
		return model.Movie{}, err
	}

	return m, nil
}

// SearchMovieByName retrieves a movie by its name
func (r *MovieRepository) SearchMovieByName(ctx context.Context, name string, order string, genre *int) ([]model.Movie, error) {
	const defaultOrderBy = "popularity DESC"
	const limit int = 10

	validOrders := map[string]string{
		"score":      "score DESC",
		"name":       "title",
		"date":       "release_year DESC",
		"popularity": defaultOrderBy,
	}
	orderBy, ok := validOrders[order]
	if !ok {
		orderBy = defaultOrderBy
		r.logger.Info("Invalid order by parameter, defaulting to popularity", common.Envelop{"provided_order": order})
	}

	op := "store.SearchMovieByName"
	meta := common.Envelop{
		"search_query": name,
		"order_by":     orderBy,
		"genre":        genre,
		"limit":        limit,
		"context":      op,
	}

	// Get the base query string using the key
	baseQuery, err := getQuery(QuerySearchMovieByName, r.logger, meta)
	if baseQuery == "" {
		return nil, err
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(baseQuery)

	args := make([]any, 0, 3)
	args = append(args, "%"+name+"%") // $1
	argPos := 2

	if genre != nil {
		queryBuilder.WriteString(fmt.Sprintf(` AND EXISTS (
					SELECT 1 FROM movie_genres
					WHERE movie_id = movies.id AND genre_id = $%d
				)`, argPos))
		args = append(args, *genre)
		argPos++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s LIMIT $%d", orderBy, argPos)) // Added space before ORDER BY
	args = append(args, limit)

	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []model.Movie{}, nil
		}
		meta["db_error_type"] = "query"
		r.logger.Errorf("Failed to query movies by name", err, meta)
		return nil, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}

	defer rows.Close()

	movies := make([]model.Movie, 0, limit)
	for rows.Next() {
		var m model.Movie
		err := rows.Scan(
			&m.ID, &m.TMDB_ID, &m.Title, &m.Tagline, &m.ReleaseYear, &m.Overview, &m.Score, &m.Popularity, &m.Language, &m.PosterURL, &m.TrailerURL, // FIXED: TMDB_ID
		)
		if err != nil {
			meta["db_error_type"] = "query_scan"
			r.logger.Errorf("Failed to scan row into movie model", err, meta)
			return nil, apperror.ErrDatabaseOpFailed(
				apperror.CodeDatabaseError,
				apperror.ErrQueryFailedMsg,
				op,
				err,
				r.logger,
				meta,
			)
		}

		movies = append(movies, m)
	}

	// Check for any error during iteration
	if err = rows.Err(); err != nil {
		meta["db_error_type"] = "query_iteration"
		r.logger.Errorf("Error during rows iteration", err, meta)
		return nil, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}
	return movies, nil
}

// GetAllGenres retrieves all genres
func (r *MovieRepository) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	op := "store.GetAllGenres"
	meta := common.Envelop{
		"resource": "genres",
		"context":  op,
	}
	rows, err := r.db.Query(ctx, `SELECT id, name FROM genres ORDER BY id`)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []model.Genre{}, nil
		}
		r.logger.Errorf("Failed to query all genres", err, meta)
		meta["db_error_type"] = "query"
		return nil, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			meta,
		)
	}

	defer rows.Close()

	var genres []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			meta["db_error_type"] = "query_scan"
			r.logger.Errorf("Failed to scan row into genre model", err, meta)
			return nil, apperror.ErrDatabaseOpFailed(
				apperror.CodeDatabaseError,
				apperror.ErrQueryFailedMsg,
				op, // FIXED: Corrected to 'op'
				err,
				r.logger,
				meta,
			)

		}
		genres = append(genres, g)
	}
	return genres, nil
}

// FetchMovieRelations retrieves actors and genres related to the movie
func (r *MovieRepository) FetchMovieRelations(ctx context.Context, m *model.Movie) error {
	var (
		g         errgroup.Group
		genreMu   sync.Mutex
		actorMu   sync.Mutex
		keywordMu sync.Mutex
		movieID   = m.ID
	)

	// Fetch Genres
	g.Go(func() error {
		rows, err := r.queryRowsWithErrorHandling(ctx, QueryGetGenres, "FetchMovieRelations.Genre", nil, movieID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var local []model.Genre
		for rows.Next() {
			var g model.Genre
			if err := rows.Scan(&g.ID, &g.Name); err != nil {
				r.logger.Errorf("Scan genre row failed", err, "movie_id", movieID)
				return err
			}
			local = append(local, g)
		}

		genreMu.Lock()
		m.Genres = append(m.Genres, local...)
		genreMu.Unlock()
		return nil
	})

	// Fetch Actors
	g.Go(func() error {
		rows, err := r.queryRowsWithErrorHandling(ctx, QueryGetActors, "FetchMovieRelations.Actors", nil, movieID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var local []model.Actor
		for rows.Next() {
			var a model.Actor
			if err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.ImageURL); err != nil { // Use a.TMDB_ID
				r.logger.Errorf("Scan actor row failed", err, "movie_id", movieID)
				return err
			}
			local = append(local, a)
		}

		actorMu.Lock()
		m.Casting = append(m.Casting, local...)
		actorMu.Unlock()
		return nil
	})

	// Fetch Keywords
	g.Go(func() error {
		rows, err := r.queryRowsWithErrorHandling(ctx, QueryGetKeywords, "FetchMovieRelations.Keywords", nil, movieID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var local []string
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err != nil {
				r.logger.Errorf("Scan keyword row failed", err, "movie_id", movieID)
				return err
			}
			local = append(local, k)
		}

		keywordMu.Lock()
		m.Keywords = append(m.Keywords, local...)
		keywordMu.Unlock()
		return nil
	})

	return g.Wait()
}

// ✨ Generic Database Helpers ✨
// getMovies is a helper method to retrieve movies
func (r *MovieRepository) getMovies(ctx context.Context, queryKey string, limit int) ([]model.Movie, error) {
	op := getOp(queryKey)
	meta := common.Envelop{"limit": limit, "context": op, "query_key": queryKey}

	rows, err := r.queryRowsWithErrorHandling(ctx, queryKey, op, meta, limit)
	if err != nil {
		return nil, err
	}

	if rows == nil {
		return []model.Movie{}, nil
	}

	defer rows.Close()

	movies, err := scanRowsToSlice(rows, scanMovie, r.logger, op, meta, limit)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

// queryRowsWithErrorHandling retrieves multiple rows from the database, handling query lookup and initial query errors.
func (r *MovieRepository) queryRowsWithErrorHandling(
	ctx context.Context,
	queryKey string, // CHANGED: Now takes queryKey (string constant)
	op string,
	meta common.Envelop,
	args ...any,
) (pgx.Rows, error) {
	query, err := getQuery(queryKey, r.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Error(fmt.Sprintf("No rows found for %s", queryKey), err)
			return nil, nil
		}
		return nil, apperror.ErrDatabaseOpFailed(
			apperror.CodeDatabaseError,
			apperror.ErrQueryFailedMsg,
			op,
			err,
			r.logger,
			common.Envelop{"db_error_type": "query"},
		)
	}

	return rows, nil
}

// queryRowWithErrHandling retrieves a single row from the database, handling query lookup and initial query errors.
// It returns the pgx.Row for scanning and an error if the query cannot be executed.
func (r *MovieRepository) queryRowWithErrorHandling(
	ctx context.Context,
	queryKey string,
	meta common.Envelop,
	args ...any,
) (pgx.Row, error) {
	query, err := getQuery(queryKey, r.logger, meta)
	if err != nil || query == "" {
		return nil, err
	}
	// fmt.Println("QueryRow: ", query)
	row := r.db.QueryRow(ctx, query, args...)

	return row, nil
}

// func (r *MovieRepository) queryRowsCtx(
// 	ctx context.Context,
// 	queryKey string,
// 	movieID int,
// 	op string,
// 	meta common.Envelop,
// ) pgx.Rows {
// 	query, ok := Queries[queryKey]
// 	if !ok {
// 		err := apperror.QueryNotFoundError(queryKey, r.logger, meta)
// 		r.logger.Error("Query constant not found", err, meta)
// 		return &errRow{Err: err}
// 	}
// 	rows, err := r.db.Query(ctx, query, movieID)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return nil, apperror.ErrRecordNotFound(err, r.logger, common.Envelop{
// 				"queyr":    queryKey,
// 				"movie_id": movieID,
// 				"context":  "store.queryRowsCtx",
// 			})
// 		}
// 		r.logger.Error("Query failed", err, "query_key", queryKey, "movie_id", movieID)
// 		// return nil, fmt.Errorf("query [%s]: %w", queryKey, err)
// 		return nil, apperror.ErrDatabaseOpFailed(
// 			apperror.CodeDatabaseError,
// 			apperror.ErrQueryFailedMsg,
// 			"store.GetMovieByID",
// 			err,
// 			r.logger,
// 			common.Envelop{"movie_id": movieID, "db_error_type": "query"},
// 		)
// 	}
// 	return rows, nil
// }

// func (r *MovieRepository) getMovies(ctx context.Context, query string, limit int) ([]models.Movie, error) {
// 	op := fmt.Sprintf("store.%s", strings.TrimPrefix(query, "Query"))

// 	rows, err := r.db.Query(ctx, query, DEFAULT_LIMIT)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			r.logger.Error("failed to query movies", err)
// 			return []models.Movie{}, nil
// 		}
// 		return []models.Movie{}, apperror.ErrDatabaseOpFailed(
// 			apperror.CodeDatabaseError,
// 			apperror.ErrQueryFailedMsg,
// 			op,
// 			err,
// 			r.logger,
// 			common.Envelop{"db_error_type": "query"},
// 		)
// 	}

// 	defer rows.Close()

// 	movies := make([]models.Movie, 0, DEFAULT_LIMIT)
// 	for rows.Next() {
// 		var m models.Movie
// 		if err := rows.Scan(
// 			&m.ID,
// 			&m.TMDB_ID,
// 			&m.Title,
// 			&m.Tagline,
// 			&m.ReleaseYear,
// 			&m.Overview,
// 			&m.Score,
// 			&m.Popularity,
// 			&m.Language,
// 			&m.PosterURL,
// 			&m.TrailerURL,
// 		); err != nil {

// 			r.logger.Error("Failed to scan row into movie model", err, common.Envelop{"op": op, "error": err.Error(), "movie_title": &m.Title})
// 			return []models.Movie{}, apperror.ErrDatabaseOpFailed(
// 				apperror.CodeDatabaseError,
// 				apperror.ErrQueryFailedMsg,
// 				op,
// 				err,
// 				r.logger,
// 				common.Envelop{"movie_title": &m.Title, "db_error_type": "query_scan"},
// 			)
// 		}

// 		movies = append(movies, m)
// 	}

// 	// Check for any error during iteration
// 	if err = rows.Err(); err != nil {
// 		r.logger.Error("Error during rows iteration", err, common.Envelop{"op": op, "error": err.Error()})
// 		return nil, apperror.ErrDatabaseOpFailed(
// 			apperror.CodeDatabaseError,
// 			apperror.ErrQueryFailedMsg,
// 			op,
// 			err,
// 			r.logger,
// 			common.Envelop{"db_error_type": "rows_iteration"},
// 		)
// 	}
// 	return movies, nil
// }

// func (r *MovieRepository) SearchMovieByName(ctx context.Context, name string, order string, genre *int) ([]models.Movie, error) {
// 	const defaultOrderBy = "popularity DESC"

// 	validOrders := map[string]string{
// 		"score":      "score DESC",
// 		"name":       "title",
// 		"date":       "release_year DESC",
// 		"popularity": defaultOrderBy,
// 	}
// 	orderBy, ok := validOrders[order]
// 	if !ok {
// 		orderBy = defaultOrderBy
// 		r.logger.Info("Invalid order by parameter, defaulting to popularity", common.Envelop{"provided_order": order})
// 	}

// 	baseQuery, ok := Queries[QuerySearchMovieByName]
// 	if !ok {
// 		return nil, apperror.QueryNotFoundError("QuerySearchMovieByName", r.logger, common.Envelop{"movie_title": name})
// 	}

// 	var queryBuilder strings.Builder
// 	queryBuilder.WriteString(baseQuery)

// 	args := make([]any, 0, 3)
// 	args = append(args, "%"+name+"%") // %1
// 	argPos := 2

// 	if genre != nil {
// 		queryBuilder.WriteString(fmt.Sprintf(` AND EXISTS (
//             SELECT 1 FROM movie_genres
//             WHERE movie_id = movies.id AND genre_id = $%d
//         )`, argPos))
// 		args = append(args, *genre)
// 		argPos++
// 	}

// 	queryBuilder.WriteString(fmt.Sprintf("ORDER BY %s LIMIT $%d", orderBy, argPos))
// 	args = append(args, DEFAULT_LIMIT)

// 	op := "store.SearchMovieByName"
// 	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return []models.Movie{}, nil
// 		}
// 		return []models.Movie{}, apperror.ErrDatabaseOpFailed(
// 			apperror.CodeDatabaseError,
// 			apperror.ErrQueryFailedMsg,
// 			op,
// 			err,
// 			r.logger,
// 			common.Envelop{"movie_title": name, "db_error_type": "query"},
// 		)
// 	}

// 	defer rows.Close()

// 	movies := make([]models.Movie, 0, DEFAULT_LIMIT)
// 	for rows.Next() {
// 		var m models.Movie
// 		err := rows.Scan(
// 			&m.ID, &m.TMDB_ID, &m.Title, &m.Tagline, &m.ReleaseYear, &m.Overview, &m.Score, &m.Popularity, &m.Language, &m.PosterURL, &m.TrailerURL,
// 		)
// 		if err != nil {
// 			r.logger.Error("Failed to scan row into movie model", err, common.Envelop{"op": op, "error": err.Error(), "movie_title": name})
// 			return []models.Movie{}, apperror.ErrDatabaseOpFailed(
// 				apperror.CodeDatabaseError,
// 				apperror.ErrQueryFailedMsg,
// 				op,
// 				err,
// 				r.logger,
// 				common.Envelop{"movie_title": name, "db_error_type": "query_scan"},
// 			)
// 		}

// 		movies = append(movies, m)
// 	}

// 	// Check for any error during iteration
// 	if err = rows.Err(); err != nil {
// 		r.logger.Error("Error during rows iteration", err, common.Envelop{"op": op, "error": err.Error(), "movie_title": name})
// 		return nil, apperror.ErrDatabaseOpFailed(
// 			apperror.CodeDatabaseError,
// 			apperror.ErrQueryFailedMsg,
// 			op,
// 			err,
// 			r.logger,
// 			common.Envelop{"movie_title": name, "db_error_type": "rows_iteration"},
// 		)
// 	}
// 	return movies, nil
// }

// ===================================================
// func (r *MovieRepository) FetchMovieRelations(ctx context.Context, m *model.Movie) error {
// 	var (
// 		g         errgroup.Group
// 		genreMu   sync.Mutex
// 		actorMu   sync.Mutex
// 		keywordMu sync.Mutex
// 		movieID   = m.ID
// 	)

// 	// Fetch Genres
// 	g.Go(func() error {
// 		rows, err := r.queryRowsCtx(ctx, QueryGetGenres, movieID)
// 		if err != nil {
// 			return err
// 		}
// 		defer rows.Close()

// 		var local []model.Genre
// 		for rows.Next() {
// 			var g model.Genre
// 			if err := rows.Scan(&g.ID, &g.Name); err != nil {
// 				r.logger.Errorf("Scan genre row failed", err, "movie_id", movieID)
// 				return err
// 			}
// 			local = append(local, g)
// 		}

// 		genreMu.Lock()
// 		m.Genres = append(m.Genres, local...)
// 		genreMu.Unlock()
// 		return nil
// 	})

// 	// Fetch Actors
// 	g.Go(func() error {
// 		rows, err := r.queryRowsCtx(ctx, QueryGetActors, movieID)
// 		if err != nil {
// 			return err
// 		}
// 		defer rows.Close()

// 		var local []model.Actor
// 		for rows.Next() {
// 			var a model.Actor
// 			if err := rows.Scan(&a.ID, &a.TMDB_ID, &a.FirstName, &a.LastName, &a.ImageURL, &a.Character); err != nil { // Use a.TMDB_ID
// 				r.logger.Errorf("Scan actor row failed", err, "movie_id", movieID)
// 				return err
// 			}
// 			local = append(local, a)
// 		}

// 		actorMu.Lock()
// 		m.Casting = append(m.Casting, local...)
// 		actorMu.Unlock()
// 		return nil
// 	})

// 	// Fetch Keywords
// 	g.Go(func() error {
// 		rows, err := r.queryRowsCtx(ctx, QueryGetKeywords, movieID)
// 		if err != nil {
// 			return err
// 		}
// 		defer rows.Close()

// 		var local []string
// 		for rows.Next() {
// 			var k string
// 			if err := rows.Scan(&k); err != nil {
// 				r.logger.Errorf("Scan keyword row failed", err, "movie_id", movieID)
// 				return err
// 			}
// 			local = append(local, k)
// 		}

// 		keywordMu.Lock()
// 		m.Keywords = append(m.Keywords, local...)
// 		keywordMu.Unlock()
// 		return nil
// 	})

// 	return g.Wait()
// }
// =============================================
// func (r *MovieRepository) queryRowsCtx(ctx context.Context, queryKey string, movieID int) (pgx.Rows, error) {
// 	op := getOp(queryKey)
// 	meta := common.Envelop{
// 		"query":    queryKey,
// 		"movie_id": movieID,
// 		"context":  op,
// 	}

// 	query, err := getQuery(queryKey, r.logger, meta)
// 	if query == "" {
// 		return nil, err
// 	}
// 	fmt.Println("Inside QueryRowsCtx: ", query)
// 	rows, err := r.queryRowsWithErrorHandling(ctx, query, op, meta, movieID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return rows, nil
// rows, err := r.db.Query(ctx, query, movieID)
// if err != nil {
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return nil, nil
// 	}
// 	meta["db_error_type"] = "query"
// 	r.logger.Errorf("Movie relation query failed", err, meta)
// 	return nil, apperror.ErrDatabaseOpFailed(
// 		apperror.CodeDatabaseError,
// 		apperror.ErrQueryFailedMsg,
// 		op,
// 		err,
// 		r.logger,
// 		meta,
// 	)
// }
// return rows, nil
// }

// ============SAVE MOVIE==================

// SaveMovie saves a movie and its related data (genres, casting, keywords) to the database.
// It handles transactions for atomicity and uniqueness where appropriate.
// It also enriches existing records by updating missing fields if new data is available.
// The search_vector column is handled by database triggers.
// func (r *MovieRepository) SaveMovie(ctx context.Context, movie *model.Movie) error {
// 	tx, err := r.db.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer func() {
// 		if recover := recover(); recover != nil {
// 			r.logger.Errorf("Panic occurred during transaction, rolling back: %v", r)
// 			tx.Rollback(ctx)
// 			panic(recover)
// 		}
// 		if err != nil {
// 			r.logger.Errorf("Error during transaction, rolling back: %v", err)
// 			tx.Rollback(ctx)
// 		}
// 	}()

// 	// 1. Insert or update the Movie
// 	insertMovieQuery := `
// 		INSERT INTO movies (tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
// 		ON CONFLICT (tmdb_id) DO UPDATE SET
// 			title = COALESCE(NULLIF(movies.title, ''), EXCLUDED.title),
// 			tagline = COALESCE(NULLIF(movies.tagline, ''), EXCLUDED.tagline),
// 			release_year = COALESCE(NULLIF(movies.release_year, 0), EXCLUDED.release_year),
// 			overview = COALESCE(movies.overview, EXCLUDED.overview),
// 			score = COALESCE(movies.score, EXCLUDED.score),
// 			popularity = COALESCE(movies.popularity, EXCLUDED.popularity),
// 			language = COALESCE(NULLIF(movies.language, ''), EXCLUDED.language),
// 			poster_url = COALESCE(movies.poster_url, EXCLUDED.poster_url),
// 			trailer_url = COALESCE(movies.trailer_url, EXCLUDED.trailer_url)
// 		RETURNING id;
// 	`
// 	var movieDBID int
// 	err = tx.QueryRow(
// 		ctx,
// 		insertMovieQuery,
// 		movie.TMDB_ID, // Use TMDB_ID here
// 		movie.Title,
// 		movie.Tagline,
// 		movie.ReleaseYear,
// 		movie.Overview,
// 		movie.Score,
// 		movie.Popularity,
// 		movie.Language,
// 		movie.PosterURL,
// 		movie.TrailerURL,
// 	).Scan(&movieDBID)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert/update movie (TMDB_ID: %d): %w", movie.TMDB_ID, err)
// 	}
// 	movie.ID = movieDBID // Update the movie struct with its new DB ID

// 	// 2. Handle Genres
// 	_, err = tx.Exec(ctx, `DELETE FROM movie_genres WHERE movie_id = $1`, movie.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to clear old movie genres for movie ID %d: %w", movie.ID, err)
// 	}
// 	for _, genre := range movie.Genres {
// 		insertGenreQuery := `INSERT INTO genres (tmdb_id, name) VALUES ($1, $2) ON CONFLICT (tmdb_id) DO UPDATE SET name = EXCLUDED.name RETURNING id;`
// 		var genreDBID int
// 		err = tx.QueryRow(ctx, insertGenreQuery, genre.ID, genre.Name).Scan(&genreDBID)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert/get genre '%s': %w", genre.Name, err)
// 		}
// 		_, err = tx.Exec(ctx, `INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, movie.ID, genreDBID)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert movie-genre association (MovieID: %d, GenreID: %d): %w", movie.ID, genreDBID, err)
// 		}
// 	}

// 	// 3. Handle Casting (Actors and Movie-Actor associations)
// 	_, err = tx.Exec(ctx, `DELETE FROM movie_actors WHERE movie_id = $1`, movie.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to clear old movie actors for movie ID %d: %w", movie.ID, err)
// 	}
// 	for _, actor := range movie.Casting {
// 		insertActorQuery := `INSERT INTO actors (tmdb_id, first_name, last_name, image_url) VALUES ($1, $2, $3, $4) ON CONFLICT (tmdb_id) DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, image_url = EXCLUDED.image_url RETURNING id;`
// 		var actorDBID int
// 		err = tx.QueryRow(ctx, insertActorQuery, actor.TMDB_ID, actor.FirstName, actor.LastName, actor.ImageURL).Scan(&actorDBID) // Use actor.TMDB_ID
// 		if err != nil {
// 			return fmt.Errorf("failed to insert/get actor '%s': %w", actor.FirstName, err)
// 		}
// 		_, err = tx.Exec(ctx, `INSERT INTO movie_actors (movie_id, actor_id, character) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`, movie.ID, actorDBID, actor.Character)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert movie-actor association (MovieID: %d, ActorID: %d): %w", movie.ID, actorDBID, err)
// 		}
// 	}

// 	// 4. Handle Keywords (assuming separate keywords and movie_keywords tables)
// 	_, err = tx.Exec(ctx, `DELETE FROM movie_keywords WHERE movie_id = $1`, movie.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to clear old movie keywords for movie ID %d: %w", movie.ID, err)
// 	}
// 	for _, keywordName := range movie.Keywords {
// 		insertKeywordQuery := `INSERT INTO keywords (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id;`
// 		var keywordDBID int
// 		err = tx.QueryRow(ctx, insertKeywordQuery, keywordName).Scan(&keywordDBID)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert/get keyword '%s': %w", keywordName, err)
// 		}
// 		_, err = tx.Exec(ctx, `INSERT INTO movie_keywords (movie_id, keyword_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, movie.ID, keywordDBID)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert movie-keyword association (MovieID: %d, KeywordID: %d): %w", movie.ID, keywordDBID, err)
// 		}
// 	}

// 	err = tx.Commit(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}
// 	return nil
// }

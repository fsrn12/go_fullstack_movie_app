package store

import (
	"fmt"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
)

type QueryKey string

// --- Query Definitions ---
// MOVIES
const (
	QueryGetTopMovies           = "GetTopMovies"
	QueryGetRandomMovies        = "RandomMovies"
	QueryGetMovieByID           = "GetMovieByID"
	QuerySearchMovieByName      = "SearchMovieByName"
	QueryGetGenreByMovieID      = "GetGenreByMovieID"
	QueryGetActorByMovieID      = "GetActorByMovieID"
	QueryGetKeywordByMovieID    = "GetKeywordByMovieID"
	QueryGetGenres              = "GetGenres"
	QueryGetActors              = "GetActors"
	QueryGetKeywords            = "GetKeywords"
	QueryGetFavorite            = "GetFavorite"
	QueryGetWatchlist           = "GetWatchlist"
	QueryIfMovieExists          = "IfMovieExists"
	QueryAddToCollection        = "AddToCollection"
	QueryRemoveFromToCollection = "RemoveFromCollection"
)

// USERS
const (
	QueryCreateUser           = "CreateUser"
	QueryGetUserByEmail       = "GetUserByEmail"
	QueryGetUserByID          = "GetUserByID"
	QueryGetUserIDByEmail     = "GetUserIDByEmail"
	QueryUpdateLastLogin      = "UpdateLastLogin"
	QueryGetUserDetails       = "GetUserDetails"
	QueryGetUserFromTokenHash = "GetUserFromTokenHash"
	QueryUpdateUserDetails    = "UpdateUserDetails"
)

// TOKENS
const (
	QuerySaveRefreshToken       = "SaveRefreshToken"
	QueryGetRefreshTokenHash    = "GetRefreshTokenHash"
	QueryDeleteRefreshToken     = "DeleteRefreshToken"
	QueryDeleteAllTokensForUser = "DeleteAllTokensForUser"
)

var Queries = map[string]string{
	// MOVIES
	QueryGetTopMovies: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
	FROM movies
	ORDER BY popularity DESC
	LIMIT $1`,

	QueryGetRandomMovies: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
	FROM movies
	ORDER BY random()
	LIMIT $1`,

	QueryGetMovieByID: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
	FROM movies
	WHERE id = $1`,

	QuerySearchMovieByName: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
	FROM movies
	WHERE (title ILIKE $1 OR overview ILIKE $1)`,

	QueryGetGenres: `SELECT g.id, g.name
	FROM genres g
	JOIN movie_genres mg ON g.id = mg.genre_id
	WHERE mg.movie_id = $1`,

	QueryGetActors: `SELECT a.id, a.first_name, a.last_name, a.image_url
	FROM actors a
	JOIN movie_cast mc ON a.id = mc.actor_id
	WHERE mc.movie_id = $1`,

	QueryGetKeywords: `SELECT k.word
	FROM keywords k
	JOIN movie_keywords mk ON k.id = mk.keyword_id
	WHERE mk.movie_id = $1`,

	// QueryGetGenreByMovieID: `SELECT g.id, g.name
	// FROM genres g
	// JOIN movie_genres mg ON g.id = mg.genre_id
	// WHERE mg.movie_id = $1`,

	// QueryGetActorByMovieID: `SELECT a.id, a.tmdb_id, a.first_name, a.last_name, a.image_url
	// FROM actors a
	// JOIN movie_actors ma ON a.id = ma.actor_id
	// WHERE ma.movie_id = $1`,

	// QueryGetKeywordByMovieID: `SELECT k.name
	// FROM keywords k
	// JOIN movie_keywords mk ON k.id = mk.keyword_id
	// WHERE mk.movie_id = $1`,

	QueryGetFavorite: `SELECT m.id, m.tmdb_id, m.title, m.tagline, m.release_year,
	m.overview, m.score, m.popularity, m.language, m.poster_url, m.trailer_url
	FROM movies m
	JOIN user_movies um ON m.id = um.movie_id
	WHERE um.user_id = $1 AND um.relation_type = 'favorite'`,

	QueryGetWatchlist: `SELECT m.id, m.tmdb_id, m.title, m.tagline, m.release_year,
	m.overview, m.score, m.popularity, m.language,
	m.poster_url, m.trailer_url
	FROM movies m
	JOIN user_movies um ON m.id = um.movie_id
	WHERE um.user_id = $1 AND um.relation_type = 'watchlist'`,

	QueryIfMovieExists: `SELECT EXISTS(
		SELECT 1
		FROM user_movies
		WHERE user_id = $1
		AND movie_id = $2
		AND relation_type = $3
	)`,

	QueryAddToCollection: `INSERT INTO user_movies (user_id, movie_id, relation_type, time_added)
	VALUES ($1, $2, $3, $4)`,

	QueryRemoveFromToCollection: `DELETE FROM user_movies
	WHERE user_id = $1 AND movie_id = $2 AND relation_type = $3`,

	// USERS
	QueryCreateUser: `INSERT INTO users (name, email, password_hashed, time_created)
	VALUES ($1, $2, $3, $4)
	RETURNING id`,

	QueryGetUserByEmail: `SELECT id, name, email, password_hashed, profile_picture_url, last_login, time_created, updated_at,
  time_deleted,time_confirmed
	FROM users
	WHERE email = $1 AND time_deleted IS NULL`,

	QueryGetUserByID: `SELECT id, name, email, password_hashed, profile_picture_url, last_login, time_created, updated_at,
  time_deleted,time_confirmed
	FROM users
	WHERE id = $1 AND time_deleted IS NULL`,

	QueryGetUserIDByEmail: `SELECT id
	FROM users
	WHERE email = $1 AND time_deleted IS NULL`,

	QueryUpdateUserDetails: `UPDATE users
	SET name = $1, email=$2, profile_picture_url = $3, updated_at=CURRENT_TIMESTAMP
	WHERE id = $4
	RETURNING updated_at`,

	QueryUpdateLastLogin: `UPDATE users
	SET last_login = $1
	WHERE id = $2`,

	QueryGetUserDetails: `SELECT id, name, email
	FROM users
	WHERE email = $1 AND time_deleted IS NULL`,

	// TOKENS
	QuerySaveRefreshToken: `INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`,

	QueryGetRefreshTokenHash: `SELECT hash FROM tokens
	WHERE hash = $1 AND scope = 'refresh' AND expiry > now()
	LIMIT 1;`,

	QueryDeleteRefreshToken: `DELETE FROM tokens
	WHERE user_id = $1 AND hash = $2 AND scope = 'refresh'`,

	QueryDeleteAllTokensForUser: `DELETE FROM tokens
	WHERE scope=$1 AND user_id=$2`,

	QueryGetUserFromTokenHash: `SELECT u.id, u.name, u.email, u.time_created, u.time_confirmed
	FROM users u
	INNER JOIN tokens t ON t.user_id=u.id
	WHERE t.hash=$1 AND t.scope=$2 AND t.expiry>$3`,
}

// getQuery retrieves a SQL query string from the Queries map.
func getQuery(q string, logger logging.Logger, meta common.Envelop) (string, error) {
	// fmt.Printf("DEBUG: getQuery - Attempting to get query for key: %s\n", q)
	query, ok := Queries[q]
	if !ok {
		logger.Errorf(fmt.Sprintf("failed to get query for %s: query not found", q), nil, meta)
		return "", apperror.QueryNotFoundError(q, logger, meta)
	}
	return query, nil
}

// // ---------------------
// // Query Keys by Domain
// // ---------------------

// // MOVIES
// const (
// 	QueryGetTopMovies        = "GetTopMovies"
// 	QueryGetRandomMovies     = "GetRandomMovies"
// 	QueryGetMovieByID        = "GetMovieByID"
// 	QuerySearchMovieByName   = "SearchMovieByName"
// 	QueryGetGenreByMovieID   = "GetGenreByMovieID"
// 	QueryGetActorByMovieID   = "GetActorByMovieID"
// 	QueryGetKeywordByMovieID = "GetKeywordByMovieID"
// 	QueryGetGenres           = "GetGenres"
// 	QueryGetActors           = "GetActors"
// 	QueryGetKeywords         = "GetKeywords"
// 	QueryGetFavorite         = "GetFavorite"
// 	QueryGetWatchlist        = "GetWatchlist"
// 	QueryIfMovieExists       = "IfMovieExists"
// 	QueryAddToCollection     = "AddToCollection"
// )

// // USERS
// const (
// 	QueryCreateUser       = "CreateUser"
// 	QueryGetUserByEmail   = "GetUserByEmail"
// 	QueryGetUserByID      = "GetUserByID"
// 	QueryGetUserIDByEmail = "GetUserIDByEmail"
// 	QueryUpdateLastLogin  = "UpdateLastLogin"
// 	QueryGetUserDetails   = "GetUserDetails"
// )

// // TOKENS
// const (
// 	QuerySaveRefreshToken       = "SaveRefreshToken"
// 	QueryGetRefreshTokenHash    = "GetRefreshTokenHash"
// 	QueryDeleteRefreshToken     = "DeleteRefreshToken"
// 	QueryDeleteAllTokensForUser = "DeleteAllTokensForUser"
// )

// // ----------------
// // SQL Query Map
// // ----------------

// var Queries = map[string]string{
// 	// MOVIES
// 	QueryGetTopMovies: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
// 	FROM movies
// 	ORDER BY popularity DESC
// 	LIMIT $1`,

// 	QueryGetRandomMovies: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
// 	FROM movies
// 	ORDER BY random()
// 	LIMIT $1`,

// 	QueryGetMovieByID: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
// 	FROM movies
// 	WHERE id = $1`,

// 	QuerySearchMovieByName: `SELECT id, tmdb_id, title, tagline, release_year, overview, score, popularity, language, poster_url, trailer_url
// 	FROM movies
// 	WHERE (title ILIKE $1 OR overview ILIKE $1)`,

// 	QueryGetGenreByMovieID: `SELECT g.id, g.name
// 	FROM genres g
// 	JOIN movie_genres mg ON g.id = mg.genre_id
// 	WHERE mg.movie_id = $1`,

// 	QueryGetActorByMovieID: `SELECT a.id, a.first_name, a.last_name, a.image_url
// 	FROM actors a
// 	JOIN movie_cast mc ON a.id = mc.actor_id
// 	WHERE mc.movie_id = $1`,

// 	QueryGetKeywordByMovieID: `SELECT k.word
// 	FROM keywords k
// 	JOIN movie_keywords mk ON k.id = mk.keyword_id
// 	WHERE mk.movie_id = $1`,

// 	QueryGetGenres: `SELECT g.id, g.name
// 	FROM genres g
// 	JOIN movie_genres mg ON g.id = mg.genre_id
// 	WHERE mg.movie_id = $1`,

// 	QueryGetActors: `SELECT a.id, a.first_name, a.last_name, a.image_url
// 	FROM actors a
// 	JOIN movie_cast mc ON a.id = mc.actor_id
// 	WHERE mc.movie_id = $1`,

// 	QueryGetKeywords: `SELECT k.word
// 	FROM keywords k
// 	JOIN movie_keywords mk ON k.id = mk.keyword_id
// 	WHERE mk.movie_id = $1`,

// 	QueryGetFavorite: `SELECT m.id, m.tmdb_id, m.title, m.tagline, m.release_year,
// 	m.overview, m.score, m.popularity, m.language, m.poster_url, m.trailer_url
// 	FROM movies m
// 	JOIN user_movies um ON m.id = um.movie_id
// 	WHERE um.user_id = $1 AND um.relation_type = 'favorite'`,

// 	QueryGetWatchlist: `SELECT m.id, m.tmdb_id, m.title, m.tagline, m.release_year,
// 	m.overview, m.score, m.popularity, m.language,
// 	m.poster_url, m.trailer_url
// 	FROM movies m
// 	JOIN user_movies um ON m.id = um.movie_id
// 	WHERE um.user_id = $1 AND um.relation_type = 'watchlist'`,

// 	QueryIfMovieExists: `SELECT EXISTS(
// 		SELECT 1
// 		FROM user_movies
// 		WHERE user_id = $1
// 		AND movie_id = $2
// 		AND relation_type = $3
// 	)`,

// 	QueryAddToCollection: `INSERT INTO user_movies (user_id, movie_id, relation_type, time_added)
// 	VALUES ($1, $2, $3, $4)`,

// 	// USERS
// 	QueryCreateUser: `INSERT INTO users (name, email, password_hashed, time_created)
// 	VALUES ($1, $2, $3, $4)
// 	RETURNING id`,

// 	QueryGetUserByEmail: `SELECT id, name, email, password_hashed
// 	FROM users
// 	WHERE email = $1 AND time_deleted IS NULL`,

// 	QueryGetUserByID: `SELECT id, name, email, password_hashed
// 	FROM users
// 	WHERE id = $1 AND time_deleted IS NULL`,

// 	QueryGetUserIDByEmail: `SELECT id
// 	FROM users
// 	WHERE email = $1 AND time_deleted IS NULL`,

// 	QueryUpdateLastLogin: `UPDATE users
// 	SET last_login = $1
// 	WHERE id = $2`,

// 	QueryGetUserDetails: `SELECT id, name, email
// 	FROM users
// 	WHERE email = $1 AND time_deleted IS NULL`,

// 	// TOKENS
// 	QuerySaveRefreshToken: `INSERT INTO tokens (hash, user_id, expiry, scope)
// 	VALUES ($1, $2, $3, $4)`,

// 	QueryGetRefreshTokenHash: `SELECT hash FROM tokens
// 	WHERE hash = $1 AND scope = 'refresh' AND expiry > now()
// 	LIMIT 1;`,

// 	QueryDeleteRefreshToken: `DELETE FROM tokens
// 	WHERE user_id = $1 AND hash = $2 AND scope = 'refresh'`,

// 	QueryDeleteAllTokensForUser: `DELETE FROM tokens
// 	WHERE scope=$1 AND user_id=$2`,
// }

// func getQuery(q string, log logging.Logger, meta common.Envelop) (string, error) {
// 	query, ok := Queries[q]
// 	if !ok {

// 		log.Errorf("failed to get query for %s: query not found", q)
// 		return "", apperror.QueryNotFoundError(q, log, meta)
// 	}
// 	return query, nil
// }

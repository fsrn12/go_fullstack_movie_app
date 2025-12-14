package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"multipass/internal/store"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
	"multipass/pkg/response"
	"multipass/pkg/utils"
)

type MovieHandler struct {
	BaseHandler
	movieStore store.MovieStore
}

func NewMovieHandler(movieStore store.MovieStore, logger logging.Logger, responder response.Writer) *MovieHandler {
	return &MovieHandler{
		movieStore: movieStore,
		BaseHandler: BaseHandler{
			Logger:       logger,
			Responder:    responder,
			ErrorHandler: apperror.NewBaseErrorHandler(logger, responder),
		},
	}
}

// GetTopMovies handles get top movies route
func (h *MovieHandler) HandleGetTopMovies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	movies, err := h.movieStore.GetTopMovies(ctx)
	if h.ErrorHandler.HandleAppError(w, r, err, "failed to fetch top movies") {
		return
	}

	resp := common.MoviesResponse{
		Success: true,
		Movies:  movies,
		Count:   len(movies),
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}

	firstMovieTitle := ""
	if len(movies) > 0 {
		firstMovieTitle = movies[0].Title
	}

	h.Logger.Info(fmt.Sprintf("GetTopMovies successfully sent top movies: first_title: %s", firstMovieTitle))
}

// GetRandomMovies handles get random movies route
func (h *MovieHandler) HandleGetRandomMovies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	movies, err := h.movieStore.GetRandomMovies(ctx)
	if h.ErrorHandler.HandleAppError(w, r, err, "failed to fetch random movies") {
		return
	}

	resp := common.MoviesResponse{
		Success: true,
		Movies:  movies,
		Count:   len(movies),
	}

	err = h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp})
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}

	firstMovieTitle := ""
	if len(movies) > 0 {
		firstMovieTitle = movies[0].Title
	}

	h.Logger.Info(fmt.Sprintf("GetRandomMovies successfully sent random movies: first_title: %s", firstMovieTitle))
}

// GetMovieByID handles get movie by ID route
func (h *MovieHandler) HandleGetMovieByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	errMeta := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
	}

	// 1: Read {ID} from request parameters
	id, err := utils.GetParamID(r)
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInvalidIDParameter(err, h.Logger, errMeta), "params_movie_id")
		return
	}

	// 2: Get movie using id
	movie, err := h.movieStore.GetMovieByID(ctx, id)
	if h.ErrorHandler.HandleAppError(w, r, err, "failed to get movie by ID") {
		return
	}

	// 3: Send Back Response
	err = h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": movie})
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer")
		return
	}

	h.Logger.Info(fmt.Sprintf("GetMovieByID successfully sent movie: %s", movie.Title))
}

// SearchMovies handles search movies by order/genre/query route
func (h *MovieHandler) HandleSearchMovies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query, err := utils.GetQueryParam(r, "q")
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, err, "missiong_query_param")
		return
	}

	if query == "" {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrBadRequest(
			errors.New("query parameters 'q' is required and cannot be empty"),
			h.Logger,
			nil,
		), "empty_empty_query")
		return
	}

	order, _ := utils.GetQueryParam(r, "order")

	errMeta := common.Envelop{
		"method":       r.Method,
		"path":         r.URL.Path,
		"order_by":     order,
		"search_query": query,
	}

	var genre *int
	genreStr, _ := utils.GetQueryParam(r, "genre")
	if genreStr != "" {
		genreInt, err := strconv.Atoi(genreStr)
		if err != nil {
			errMeta["genre_id"] = genreStr
			h.ErrorHandler.HandleAppError(w, r, apperror.ErrInvalidGenreID(err, h.Logger, errMeta), "invalid_genre_id")
			return
		}

		genre = &genreInt
	}

	movies, err := h.movieStore.SearchMovieByName(ctx, query, order, genre)
	if h.ErrorHandler.HandleAppError(w, r, err, "failed to search movies") {
		return
	}

	resp := common.MoviesResponse{
		Success: true,
		Movies:  movies,
		Count:   len(movies),
	}

	err = h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp})
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer")
		return
	}
	h.Logger.Info("SearchMovies successfully sent results")
}

// GetAllGenres handles get all genres route
func (h *MovieHandler) HandleGetAllGenres(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	genres, err := h.movieStore.GetAllGenres(ctx)
	if h.ErrorHandler.HandleAppError(w, r, err, "failed to fetch genres") {
		return
	}

	err = h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data":  genres,
		"count": len(genres),
	})
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer")
		return
	}

	h.Logger.Info("GetAllGenres successfully sent genres")
}

// writeAppError is a helper method to write error response
// func (h *MovieHandler) writeAppError(w http.ResponseWriter, r *http.Request, appErr *apperror.AppError) {
// 	if appErr == nil {
// 		// fallback in case of programming error
// 		appErr = apperror.NewAppError(
// 			apperror.CodeInternal,
// 			"Unknown application error",
// 			"writeAppError",
// 			nil,
// 			h.Logger,
// 			common.Envelop{
// 				"status": "error",
// 				"path":   r.URL.Path,
// 				"method": r.Method,
// 			},
// 		)
// 	}

// 	if appErr.Logger != nil {
// 		appErr.LogError()
// 	}

// 	appErr.WriteJSONError(w, r, h.Responder)
// }

// # sendErr is a helper method to send json writing error response
// func (h *MovieHandler) sendErr(w http.ResponseWriter, r *http.Request, err error) {
// 	errMeta := common.Envelop{
// 		"method":     r.Method,
// 		"path":       r.URL.Path,
// 		"statusCode": http.StatusInternalServerError,
// 	}
// 	resErr := apperror.ErrFailedJSONWriter(err, h.Logger, errMeta)
// 	h.Logger.Error(apperror.ErrFailedJSONResWrite, resErr, errMeta)
// 	resErr.WriteJSONError(w, r, h.Responder)
// }

// func (h *MovieHandler) handleStoreError(w http.ResponseWriter, r *http.Request, err error, context string) bool {
// 	if err == nil {
// 		return false
// 	}

// 	var status int
// 	var userMsg error

// 	switch {
// 	case errors.Is(err, utils.ErrMovieNotFound),
// 		errors.Is(err, utils.ErrNoRecordFound),
// 		errors.Is(err, utils.ErrResourceNotFound):
// 		status = http.StatusNotFound
// 		userMsg = utils.ErrResourceNotFound

// 	case errors.Is(err, utils.ErrInvalidIDParameter),
// 		errors.Is(err, utils.ErrInvalidMovieID),
// 		errors.Is(err, utils.ErrInvalidActorID),
// 		errors.Is(err, utils.ErrInvalidGenreID),
// 		errors.Is(err, utils.ErrInvalidRequestPayload),
// 		errors.Is(err, utils.ErrInvalidContentType):
// 		status = http.StatusBadRequest
// 		userMsg = err // show exact cause for client

// 	case errors.Is(err, utils.ErrUnauthorizedAccess),
// 		errors.Is(err, utils.ErrMissingToken),
// 		errors.Is(err, utils.ErrInvalidToken),
// 		errors.Is(err, utils.ErrExpiredInvalidToken),
// 		errors.Is(err, utils.ErrInvalidLogin),
// 		errors.Is(err, utils.ErrInvalidAuthHeader),
// 		errors.Is(err, utils.ErrInvalidTokenFmt):
// 		status = http.StatusUnauthorized
// 		userMsg = utils.ErrUnauthorizedAccess

// 	case errors.Is(err, utils.ErrForbiddenAccess),
// 		errors.Is(err, utils.ErrAdminOnly):
// 		status = http.StatusForbidden
// 		userMsg = utils.ErrForbiddenAccess

// 	case errors.Is(err, utils.ErrFailedToCreate),
// 		errors.Is(err, utils.ErrFailedToUpdate),
// 		errors.Is(err, utils.ErrFailedToDelete):
// 		status = http.StatusUnprocessableEntity
// 		userMsg = err // pass original error to client

// 	default:
// 		status = http.StatusInternalServerError
// 		userMsg = utils.ErrInternalServer
// 	}

// 	h.logger.Error("Request failed",
// 		err,
// 		"context", context,
// 		"method", r.Method,
// 		"path", r.URL.Path,
// 		"status", status,
// 	)

// 	utils.NewAPIError(status, userMsg.Error()).ErrResponse(w, h.responder)

// 	return true
// }

// var placeholderFilms = []models.Movie{
// 	{
// 		ID:          1,
// 		TMDB_ID:     101,
// 		Title:       "The Hacker",
// 		ReleaseYear: 2022,
// 		Genres:      []models.Genre{{ID: 1, Name: "Thriller"}},
// 		Keywords:    []string{"hacking", "cybercrime"},
// 		Casting:     []models.Actor{{ID: 1, FirstName: "Jane", LastName: "Dalia"}},
// 	},
// 	{
// 		ID:          2,
// 		TMDB_ID:     102,
// 		Title:       "Space Dreams",
// 		ReleaseYear: 2020,
// 		Genres:      []models.Genre{{ID: 2, Name: "Sci-Fi"}},
// 		Keywords:    []string{"space", "exploration"},
// 		Casting:     []models.Actor{{ID: 2, FirstName: "John", LastName: "Space"}},
// 	},
// 	{
// 		ID:          3,
// 		TMDB_ID:     103,
// 		Title:       "The Lost City",
// 		ReleaseYear: 2019,
// 		Genres:      []models.Genre{{ID: 3, Name: "Adventure"}},
// 		Keywords:    []string{"jungle", "treasure"},
// 		Casting:     []models.Actor{{ID: 3, FirstName: "Lara", LastName: "Hunt"}},
// 	},
// }

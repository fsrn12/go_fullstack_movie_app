package service

import (
	"context"

	"multipass/internal/model"
	"multipass/internal/store"
	"multipass/pkg/logging"
)

type MovieService interface {
	MovieSearch(ctx context.Context, query string, order string, genre *int) ([]model.Movie, error)
}

type movieService struct {
	store  store.MovieStore
	logger logging.Logger
}

func NewMovieServie(ms store.MovieStore, logger logging.Logger) MovieService {
	return &movieService{
		store:  ms,
		logger: logger,
	}
}

// SearchMovies implements the business logic for movie search.
func (s *movieService) MovieSearch(ctx context.Context, query string, order string, genre *int) ([]model.Movie, error) {
	// Example: Add business logic here
	// Maybe log search queries for analytics:
	// s.analyticsClient.LogSearch(ctx, query) // If you add an analytics dependency

	// Call the store layer for data
	movies, err := s.store.SearchMovieByName(ctx, query, order, genre)
	if err != nil {
		// Business logic for handling store errors could go here if needed,
		// but typically, the service passes through data errors or translates them.
		return nil, err
	}

	// Example: Apply post-retrieval business logic
	// e.g., Filter movies based on user preferences, add calculated fields
	// filteredMovies := s.filterSensitiveContent(movies, ctx.Value("user_id"))

	return movies, nil // Or filteredMovies
}

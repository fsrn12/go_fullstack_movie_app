package model

type Movie struct {
	ID          int      `json:"id"`
	TMDB_ID     int      `json:"tmdb_id"`
	Title       string   `json:"title"`
	Tagline     string   `json:"tagline"`
	ReleaseYear int      `json:"release_year"`
	Overview    *string  `json:"overview"`
	Score       *float32 `json:"score"`
	Popularity  *float32 `json:"popularity"`
	Language    *string  `json:"language"`
	PosterURL   *string  `json:"poster_url"`
	TrailerURL  *string  `json:"trailer_url"`
	Casting     []Actor  `json:"casting"`
	Genres      []Genre  `json:"genres"`
	Keywords    []string `json:"keywords"`
}

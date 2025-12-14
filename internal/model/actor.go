package model

import "fmt"

type Actor struct {
	ID int `json:"id"`
	// TMDB_ID   int     `json:"tmdb_id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	ImageURL  *string `json:"image_url"`
	// Character string  `json:"character"`
}

func (a Actor) Name() string {
	return fmt.Sprintf("%s %s", a.FirstName, a.LastName)
}

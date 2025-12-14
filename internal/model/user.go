package model

import "time"

type User struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	PasswordHashed    string     `json:"-"`
	IsVerified        bool       `json:"is_verified" db:"is_verified"`
	VerificationToken string     `json:"-" db:"verification_token"`
	ResetToken        string     `json:"-" db:"reset_token"`
	TokenExpiresAt    *time.Time `json:"-" db:"token_expires_at"`
	CreatedAt         time.Time  `json:"time_created"`
	UpdatedAt         time.Time  `json:"time_updated"`
	ConfirmedAt       *time.Time `json:"time_confirmed,omitempty"`
	DeletedAt         *time.Time `json:"time_deleted,omitempty"`
	LastLogin         *time.Time `json:"lastLogin,omitempty"`
	ProfilePictureURL *string    `json:"profilePictureUrl,omitempty"`
	Favorites         []Movie    `json:"favorites"`
	Watchlist         []Movie    `json:"watchlist"`
}

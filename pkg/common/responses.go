package common

import (
	"mime/multipart"
	"time"

	"multipass/internal/model"
)

type SendVerificationEmailRequest struct {
	Email string `json:"email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type SendPasswordResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type VerifyOTPRequest struct {
	Code    string `json:"code"`
	Purpose string `json:"purpose"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type RegisterResult struct {
	User         *model.User `json:"-"`
	JWT          string      `json:"jwt,omitempty"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

type AuthResult struct {
	User         *model.User `json:"-"`
	JWT          string      `json:"jwt,omitempty"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	JWT     string      `json:"jwt,omitempty"`
	User    *model.User `json:"user"`
}

type UserUpdateRequest struct {
	Name              *string `json:"name,omitempty"`
	Email             *string `json:"email,omitempty"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
}
type ProfileUpdateRequest struct {
	Name              *string
	Email             *string
	ProfilePictureURL *multipart.File
}

type UserUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UserProfileResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	User    *model.User `json:"user"`
}

type AuthRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type RefreshCookie struct {
	RefreshToken *string `json:"refresh_token"`
}

type WebAuthnAuthRequest struct {
	Email string `json:"email"`
}

type WebAuthnAuthResponse struct {
	Success bool   `json:"success"`
	JWT     string `json:"jwt"`
}

type RefreshResponse struct {
	JWT string `json:"jwt,omitempty"`
}

type ProfilePictureUploadResponse struct {
	Success           bool   `json:"success"`
	ProfilePictureURL string `json:"profilePictureUrl"`
}

type CollectionRequest struct {
	MovieID    *int    `json:"movie_id"`
	Collection *string `json:"collection"`
}

type CollectionInput struct {
	MovieID    int    `json:"movie_id"`
	Collection string `json:"collection"`
}

type CollectionSuccess struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

type MoviesResponse struct {
	Success bool          `json:"success,omitempty"`
	Movies  []model.Movie `json:"movies,omitempty"`
	Count   int           `json:"count,omitempty"`
}

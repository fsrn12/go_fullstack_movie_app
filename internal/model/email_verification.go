package model

import "time"

type EmailVerification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordReset struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type OTPVerification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Code      string    `json:"code"`
	Purpose   string    `json:"purpose"` // "login", "2fa", "sensitive_action"
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `json:"verified"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

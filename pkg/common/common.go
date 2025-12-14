package common

import (
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
)

type Error interface {
	Error() string
}

type PanicMiddleware interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler)
	PanicRecoverMiddleWare(next http.Handler) http.Handler
}

type Envelop map[string]any

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type UserContext struct {
	UserID int
	Email  string
	Name   string
}

type WebAuthnSignUpResult struct {
	Options *protocol.CredentialCreation
	Token   string
}

type WebAuthnAuthStartResult struct {
	Options *protocol.CredentialAssertion
	Token   string
}

type WebAuthnAuthenticationEndResult struct {
	JWT   string
	Token string
}

type PasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type SendOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type OTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

const (
	AllowedExtensions string = "jpg|jpeg|png"
	MaxFileSize       int64  = 10 * 1024 * 1024 // 10 MB
)

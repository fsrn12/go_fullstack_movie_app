package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"time"

	"multipass/internal/model"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	ScopeAuth              string = "authentication"
	AccessToken                   = TokenType("access")
	RefreshToken                  = TokenType("refresh")
	EmailVerificationScope string = "verification"
	PasswordResetScope     string = "reset"
	OTPScope               string = "otp"
	// EmailVerification         = TokenType("email_verification")
	// PasswordReset             = TokenType("password_reset")
	RefreshTokenLength int = 32
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int       `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type CustomClaims struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Logger        logging.Logger
}

func NewTokenManager(access, refresh string, accessTTL time.Duration, refreshTTL time.Duration, logger logging.Logger) *TokenManager {
	return &TokenManager{
		AccessSecret:  []byte(access),
		RefreshSecret: []byte(refresh),
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
		Logger:        logger,
	}
}

// CreateJWT generates a new JWT access token for a given user.
func (tm *TokenManager) CreateJWT(user *model.User) (string, error) {
	claims := CustomClaims{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.AccessTTL)), // 15m
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign new token with secret
	tokenString, err := jwtToken.SignedString(tm.AccessSecret)
	if err != nil {
		tm.Logger.Error("Failed to sign JWT access token", err)
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// CreateRefreshToken generates a new cryptographically secure opaque refresh token.
func (tm *TokenManager) CreateRefreshToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	randomBytes := make([]byte, RefreshTokenLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		tm.Logger.Error("Failed to generate random bytes for refresh token", err)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Base32 encoding makes the token URL-safe and human-readable
	plainText := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Hash the plaintext token for secure storage and comparison
	hash := sha256.Sum256([]byte(plainText))

	token := &Token{
		Plaintext: plainText,
		Hash:      hash[:],
		UserID:    int(userID),
		Expiry:    time.Now().Add(tm.RefreshTTL),
		Scope:     scope,
	}

	return token, nil
}

// ValidateJWT parses and validates a JWT access token.
func (tm *TokenManager) ValidateJWT(tokenStr string) (*CustomClaims, error) {
	op := "tokens.ValidateJWT"
	metadata := common.Envelop{
		"op": op,
	}
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (any, error) {
			// Verify Token Signing Method i.e. HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				metadata["details"] = apperror.ErrInvalidTokenSignatureMsg
				sigErr := apperror.ErrInvalidTokenSignature(nil, tm.Logger, metadata)
				tm.Logger.Errorf("Unexpected token signing method", sigErr, metadata)
				return nil, sigErr
			}
			return tm.AccessSecret, nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			metadata["details"] = "jwt_token_expired"
			tkErr := apperror.ErrTokenExpired(nil, tm.Logger, metadata)
			tm.Logger.Error("Access token expired", tkErr, metadata)
			return nil, tkErr
		}

		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			metadata["details"] = "jwt_token_invalid"
			tkErr := apperror.ErrInvalidJWT(nil, tm.Logger, metadata)
			tm.Logger.Error("Malformed or invalid JWT signature", tkErr, metadata)
			return nil, tkErr
		}

		metadata["details"] = "invalid_auth: unauthorized or suspicious activity"
		tkErr := apperror.ErrInvalidAuth(nil, tm.Logger, metadata)
		tm.Logger.Error("access token parsing error", tkErr, metadata)
		return nil, tkErr
	}

	if !token.Valid {
		metadata["details"] = fmt.Sprintf("%s: invalid_token: unauthorized or suspicious activity", apperror.ErrInvalidTokenMsg)
		tkErr := apperror.ErrInvalidJWT(nil, tm.Logger, metadata)
		tm.Logger.Error("invalid access token", tkErr, metadata)
		return nil, tkErr
	}

	return claims, nil
}

// VerifyOpaqueRefreshToken compares the plaintext refresh token with its stored hash.
// This function strictly verifies the token's integrity, not its claims or expiry.
func (tm *TokenManager) VerifyRefreshToken(plainText string, hashed []byte) error {
	if plainText == "" || len(hashed) == 0 {
		return errors.New("invalid input: plaintext or hashed token missing")
	}

	// Calculate the hash of the provided plaintext token
	calculateHash := sha256.Sum256([]byte(plainText))

	// using constant time comparison to avoid timing attacks
	if !hmac.Equal(calculateHash[:], hashed) {
		rtErr := apperror.ErrTokenMalformed(nil, tm.Logger, common.Envelop{"error": "token mismatch"})
		tm.Logger.Error("Refresh token has does not match", rtErr)
		return rtErr
	}

	return nil
}

func (tm *TokenManager) GenerateTokenPair(user *model.User) (accessToken string, refreshToken *Token, err error) {
	// ACCESS TOKEN
	accessToken, err = tm.CreateJWT(user)
	if err != nil {
		tm.Logger.Error("Failed to create access token during pair generation", err)
		return "", nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// REFRESH TOKEN
	refreshToken, err = tm.CreateRefreshToken(user.ID, tm.RefreshTTL, string(RefreshToken))
	if err != nil {
		tm.Logger.Error("Failed to create refresh token during pair generation", err)
		return "", nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (tm *TokenManager) LooksLikeToken(token string) bool {
	if len(token) != (RefreshTokenLength*8+4)/5 {
		return false
	}
	_, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(token)
	return err == nil
}

// func (tm *TokenManager) VerifyAndValidateRefreshToken(plainText string, dbHash []byte) (*CustomClaims, error) {
// 	// 1: Verify integrity
// 	if err := tm.VerifyRefreshToken(plainText, dbHash); err != nil {
// 		return nil, err
// 	}

// 	// 2: Validate structure + claims
// 	claims, err := tm.ValidateAccessToken(plainText)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return claims, nil
// }

// DecodeAccessToken decodes and validates the access token
// func (tm *TokenManager) DecodeAccessToken(tokenStr string) (*CustomClaims, error) {
// 	// 1: Validate Token
// 	claims, err := tm.ValidateToken(tokenStr, false)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return claims, nil
// }

package config

import (
	"fmt"
	"os"
	"time"

	"multipass/pkg/utils"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string           `mapstructure:"database"`
	RedisURL           string           `mapstructure:"redis_url"`
	JWT                *JWTConfig       `mapstructure:"-"`
	STATIC             string           `mapstructure:"static"`
	LogFilePath        string           `mapstructure:"log_file_path"`
	ProfilePicturePath string           `mapstructure:"profile_picture_path"`
	ProfilePictureBase string           `mapstructure:"profile_picture_base"`
	WebAuthn           *webauthn.Config `mapstructure:"webauthn"`
	Email              *EMAILConfig     `mapstructure:"email"`
}

type JWTConfig struct {
	AccessTokenSecret  string        `mapstructure:"access_token_secret"`
	RefreshTokenSecret string        `mapstructure:"refresh_token_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_expiry_minutes"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_expiry_hours"`
}

type EMAILConfig struct {
	FromAddress string `json:"from_email"`
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    string `json:"smtp_port"`
	SMTPUser    string `json:"smtp_user"`
	SMTPPass    string `json:"smtp_pass"`
	FrontendURL string `json:"base_url"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("WARNING: No .env file found or error loading .env: %v. Proceeding with environment variables and config file.\n", err)
	}

	// Logger File Path
	logFile := os.Getenv("LOG_FILE_PATH")
	if logFile == "" {
		return nil, fmt.Errorf("LOG_FILE_PATH not set in environment variables or .env file")
	}

	// Database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in environment variables or .env file")
	}

	// Redis URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL not set in environment variables or .env file")
	}

	// STATIC FILES PATH
	staticPath := os.Getenv("STATIC")
	if staticPath == "" {
		return nil, fmt.Errorf("STATIC path not set in environment variables or .env file")
	}

	// User Profile Picture Path
	profilePicturePath := os.Getenv("PROFILE_PICTURE_PATH")
	if profilePicturePath == "" {
		return nil, fmt.Errorf("PROFILE_PICTURE_PATH not set in environment variables or .env file")
	}

	// User Profile Picture Base
	profilePictureBase := os.Getenv("PROFILE_PICTURE_BASE")
	if profilePictureBase == "" {
		return nil, fmt.Errorf("PROFILE_PICTURE_BASE not set in environment variables or .env file")
	}

	// TOKENS
	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET not set in environment variables or .env file")
	}

	refreshSecret := os.Getenv("REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, fmt.Errorf("REFRESH_SECRET not set in environment variables or .env file")
	}

	accessTokenTTL := os.Getenv("ACCESS_TOKEN_TTL")
	if accessTokenTTL == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN_TTL not set in environment variables or .env file")
	}
	refreshTokenTTL := os.Getenv("REFRESH_TOKEN_TTL")
	if refreshTokenTTL == "" {
		return nil, fmt.Errorf("REFRESH_TOKEN_TTL not set in environment variables or .env file")
	}

	rpDisplayName := os.Getenv("WEBAUTHN_RP_DISPLAY_NAME")
	if rpDisplayName == "" {
		return nil, fmt.Errorf("WEBAUTHN_RP_DISPLAY_NAME not set in environment variables or .env file")
	}

	// Email Service Coinfig
	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		return nil, fmt.Errorf("EMAIL_FROM not set in environment")
	}
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return nil, fmt.Errorf("SMTP_HOST not set in environment")
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		return nil, fmt.Errorf("SMTP_PORT not set in environment")
	}
	smtpUser := os.Getenv("SMTP_USER")
	if smtpUser == "" {
		return nil, fmt.Errorf("SMTP_USER not set in environment")
	}
	smtpPass := os.Getenv("SMTP_PASS")
	if smtpPass == "" {
		return nil, fmt.Errorf("SMTP_PASS not set in environment")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL not set in environment")
	}

	emailConfig := &EMAILConfig{
		FromAddress: emailFrom,
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		SMTPUser:    smtpUser,
		SMTPPass:    smtpPass,
		FrontendURL: frontendURL,
	}

	// WebAuthn config
	rpID := os.Getenv("WEBAUTHN_RP_ID")
	if rpID == "" {
		return nil, fmt.Errorf("WEBAUTHN_RP_ID not set in environment variables or .env file")
	}

	rpOrigins := os.Getenv("WEBAUTHN_RP_ORIGINS")
	if rpOrigins == "" {
		return nil, fmt.Errorf("WEBAUTHN_RP_ORIGINS not set in environment variables or .env file")
	}

	webAuthnConfig := &webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          rpID,
		RPOrigins:     []string{rpOrigins},
	}

	jwt := &JWTConfig{
		AccessTokenSecret:  jwtAccessSecret,
		RefreshTokenSecret: refreshSecret,
		AccessTokenTTL:     utils.MustParseDuration(accessTokenTTL, 15*time.Minute),
		RefreshTokenTTL:    utils.MustParseDuration(refreshTokenTTL, 2*24*time.Hour),
	}

	return &Config{
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		JWT:                jwt,
		STATIC:             staticPath,
		LogFilePath:        logFile,
		ProfilePicturePath: profilePicturePath,
		ProfilePictureBase: profilePictureBase,
		WebAuthn:           webAuthnConfig,
		Email:              emailConfig,
	}, nil
}

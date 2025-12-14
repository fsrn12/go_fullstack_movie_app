package utils

import (
	"encoding/base32"
	"os"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
)

func ReadSecret(key string, logger logging.Logger) string {
	jwtSecret := os.Getenv(key)
	if jwtSecret == "" {
		jwtSecret = "dutHgCutnW3Ra7xmfOoaqDIVClzsP7/u7b0RnIvD46BCzpLLB"
		tokenErr := apperror.ErrMissingJWTSecret(nil, logger, common.Envelop{"error": apperror.ErrInternalServerMsg})
		logger.Error("JWT_SECRET not set, using default dev secret", tokenErr)
	} else {
		logger.Info("Using JWT_SECRET from environment")
	}

	return jwtSecret
}

func LooksLikeToken(token string) bool {
	if len(token) != 52 {
		return false
	}
	_, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(token)
	return err == nil
}

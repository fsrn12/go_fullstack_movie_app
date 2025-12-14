package hashing

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func SetHash(plainTextPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	return string(hash), nil
}

func IsPasswordMatch(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

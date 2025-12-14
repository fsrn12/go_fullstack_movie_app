package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
)

/*
## ✅ Password Validation Requirements
* Minimum 8 characters
* At least one uppercase letter
* At least one lowercase letter
* At least one number
* At least one special character (e.g., `!@#$%^&*+-_`)
* Optional: prevent commonly used passwords (dictionary check)
*/

var (
	uppercase   = regexp.MustCompile(`[A-Z]`)
	lowercase   = regexp.MustCompile(`[a-z]`)
	digit       = regexp.MustCompile(`[0-9]`)
	specialChar = regexp.MustCompile(`[!"£$%^&*()_+=\-@#:;',.?~]`)
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	nameRegex   = regexp.MustCompile(`^[A-Za-z]+([-' ][A-Za-z]+){0,2}$`)
	// nameRegex  = regexp.MustCompile(`^[A-Za-z]+(?: [A-Za-z]+){1,2}$`)
)

const passwordSpecialChars = `!"£$%^&*()_+=\-@#:;',.?~`

func ValidatePassword(password string) error {
	msg := `Password must have:
 - at least 8 characters.
 - at least one uppercase letter: [A-Z].
 - at least one lowercase letter: [a-z].
 - at least one digit: [0-9].
 - at least one special character from: [!"£$%^&*()_+=-@#:;',.?~].
 - maximum 32 characters.
 `

	const minLength = 8
	const maxLength = 32

	passwordLength := len(password)
	if passwordLength < minLength || passwordLength > maxLength {
		return fmt.Errorf("password validation failed:%s", msg)
	}

	var hasLower, hasUpper, hasDigit bool
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit {
		return fmt.Errorf("password validation failed:%s", msg)
	}

	// 3. Check for at least one special character.
	// `strings.ContainsAny` is highly optimized for this, performing a fast scan.
	hasSpecial := strings.ContainsAny(password, passwordSpecialChars)

	if !hasSpecial {
		return fmt.Errorf("password validation failed:%s", msg)
	}

	return nil
}

/*
▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩
#     REGISTER USER VALIDATION
▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩
*/

func IsValidToProcess(v *string, fieldName string, errMsg common.Envelop) (string, error) {
	if v == nil {
		return "", apperror.ErrMissingRequiredField(fieldName, nil, nil, errMsg)
	}

	value := strings.TrimSpace(*v)
	if value == "" {
		return "", apperror.ErrMissingRequiredField(fieldName, nil, nil, errMsg)
	}

	return value, nil
}

func ValidateName(n *string) (string, error) {
	name, err := IsValidToProcess(n, "Name", common.Envelop{"error": apperror.ErrInvalidNameMsg})
	if err != nil {
		return "", err
	}

	nameLength := len(name)
	if nameLength < 2 || nameLength > 32 {
		return "", apperror.ErrRequestValidation(nil, nil, common.Envelop{"field": "name", "message": apperror.ErrInvalidNameMsg})
	}

	if !nameRegex.MatchString(name) {
		return "", apperror.ErrRequestValidation(nil, nil, common.Envelop{"field": "name", "message": apperror.ErrInvalidNameMsg})
	}

	return name, nil
}

func ValidateEmail(e *string) (string, error) {
	email, err := IsValidToProcess(e, "Email", common.Envelop{"error": apperror.ErrInvalidEmailFormatMsg})
	if err != nil {
		return "", err
	}

	if !emailRegex.MatchString(email) {
		return "", apperror.ErrRequestValidation(nil, nil, common.Envelop{"field": "email", "message": apperror.ErrInvalidEmailFormatMsg})
	}

	return email, nil
}

func ValidateRegisterRequest(req *common.RegisterRequest) (common.RegisterInput, error) {
	// NAME
	name, err := ValidateName(req.Name)
	if err != nil {
		return common.RegisterInput{}, err
	}

	// EMAIL
	email, err := ValidateEmail(req.Email)
	if err != nil {
		return common.RegisterInput{}, err
	}

	if !emailRegex.MatchString(email) {
		return common.RegisterInput{}, apperror.ErrRequestValidation(nil, nil, common.Envelop{"field": "email", "message": apperror.ErrInvalidEmailFormatMsg})
	}

	// PASSWORD
	password, err := IsValidToProcess(req.Password, "Password", common.Envelop{"error": apperror.ErrInvalidPasswordMsg})
	if err != nil {
		return common.RegisterInput{}, err
	}

	if err := ValidatePassword(password); err != nil {
		return common.RegisterInput{}, apperror.ErrRequestValidation(err, nil, common.Envelop{"field": "password"})
	}

	return common.RegisterInput{Name: name, Email: email, Password: password}, nil
}

/*
▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩
#    LOGIN REQUEST VALIDATION
▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩▩
*/

func SanitizeLoginRequest(req *common.AuthRequest) (common.LoginInput, error) {
	// EMAIL
	email, err := IsValidToProcess(req.Email, "Email", common.Envelop{"error": apperror.ErrInvalidEmailFormatMsg})
	if err != nil {
		return common.LoginInput{}, err
	}

	// PASSWORD
	password, err := IsValidToProcess(req.Password, "Password", common.Envelop{"error": apperror.ErrInvalidPasswordMsg})
	if err != nil {
		return common.LoginInput{}, err
	}

	return common.LoginInput{
		Email:    email,
		Password: password,
	}, nil
}

func SanitizeCollectionReq(req *common.CollectionRequest) (common.CollectionInput, error) {
	if req.MovieID == nil {
		return common.CollectionInput{}, apperror.ErrMissingRequiredField("MovieID", nil, nil, nil)
	}

	if *req.MovieID <= 0 {
		return common.CollectionInput{}, apperror.ErrMissingRequiredField("MovieID", nil, nil, common.Envelop{"detail": "movie_id out of bound", "movie_id": *req.MovieID})
	}

	if req.Collection == nil {
		return common.CollectionInput{}, apperror.ErrMissingRequiredField("Collection", nil, nil, nil)
	}

	collection := strings.TrimSpace(*req.Collection)

	if collection != "favorite" && collection != "watchlist" {
		return common.CollectionInput{}, fmt.Errorf("invalid collection type, collection must be 'favorite' or 'watchlist'")
	}

	return common.CollectionInput{
		MovieID:    *req.MovieID,
		Collection: collection,
	}, nil
}

// SanitizeRefreshCookie checks for empty value (token string) (string, error)
func SanitizeRefreshCookie(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("refresh token cannot be empty")
	}
	tkStr := strings.TrimSpace(token)
	return tkStr, nil
}

// func SanitizeRefreshRequest(req *common.RefreshRequest) (common.RefreshData, error) {
// 	accessTokenStr, err := IsValidToProcess(req.AccessToken, "AccessToken", common.Envelop{"error": apperror.ErrInvalidEmailFormatMsg})
// 	if err != nil {
// 		return common.RefreshData{}, err
// 	}
// 	refreshTokenStr, err := IsValidToProcess(req.RefreshToken, "RefreshToken", common.Envelop{"error": apperror.ErrInvalidEmailFormatMsg})
// 	if err != nil {
// 		return common.RefreshData{}, err
// 	}

// 	return common.RefreshData{
// 		AccessToken:  accessTokenStr,
// 		RefreshToken: refreshTokenStr,
// 	}, nil
// }

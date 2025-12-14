package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"multipass/config"
	"multipass/internal/auth/hashing"
	"multipass/internal/auth/tokens"
	"multipass/internal/model"

	"multipass/internal/store"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/ctxutils"
	"multipass/pkg/imgutils"
	"multipass/pkg/logging"
	"multipass/pkg/utils"
	"multipass/pkg/validator"

	"github.com/disintegration/imaging"
)

type UserAccountService interface {
	RegisterUser(ctx context.Context, req *common.RegisterRequest) (*common.RegisterResult, error)
	AuthService(ctx context.Context, req *common.AuthRequest) (*common.AuthResult, error)
	LogoutService(ctx context.Context, refreshTokenPlaintext string) error
	RefreshService(ctx context.Context, refreshTokenPlaintext string) (*common.AuthResult, error)
	AddToCollection(ctx context.Context, req *common.CollectionRequest) (*common.CollectionSuccess, error)
	RemoveFromCollection(ctx context.Context, req *common.CollectionRequest) (*common.CollectionSuccess, error)
	AccountDetailsService(ctx context.Context, email string) (*model.User, error)
	DeleteTokenService(ctx context.Context, id int) error
	UserUpdateService(ctx context.Context, userID int, req *common.UserUpdateRequest) error
	UploadProfilePictureService(ctx context.Context, userID int, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
	VerifyEmail(ctx context.Context, plainTextToken string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, plainTextToken, newPassword string) error
}

type AccountService struct {
	BaseService
	store       store.AccountStore
	tokenStore  store.TokenStore
	tokens      *tokens.TokenManager
	emailSender EmailSender
	logger      logging.Logger
	config      *config.Config
}

func NewAccountService(
	accountStore store.AccountStore,
	tokenStore store.TokenStore,
	tokenManager tokens.TokenManager,
	emailSender EmailSender,
	logger logging.Logger,
	config *config.Config,
) *AccountService {
	return &AccountService{
		store:       accountStore,
		emailSender: emailSender,
		tokenStore:  tokenStore,
		tokens:      &tokenManager,
		logger:      logger,
		config:      config,
	}
}

// RegisterUser handles the business logic for user registration.
func (s *AccountService) RegisterUser(ctx context.Context, req *common.RegisterRequest) (*common.RegisterResult, error) {
	metaData := common.Envelop{
		"op": "service.RegisterUser",
	}

	// Sanitize and Validate Request
	registerData, err := validator.ValidateRegisterRequest(req)
	if err != nil {
		return nil, apperror.ErrBadRequest(err, s.Logger, metaData)
	}

	metaData["email"] = registerData.Email
	// Check email uniqueness
	exists, err := s.store.CheckEmailExists(ctx, registerData.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		metaData["context"] = "Duplicate Key"
		return nil, apperror.ErrDuplicateEntry(nil, s.Logger, metaData)
	}

	// Hash Password
	passwordHash, err := hashing.SetHash(registerData.Password)
	if err != nil {
		metaData["context"] = "Hashing"
		return nil, apperror.ErrInternalServer(err, s.Logger, metaData)
	}

	// Create user in database
	userID, err := s.store.CreateUser(ctx, registerData.Name, registerData.Email, passwordHash)
	if err != nil {
		return nil, err
	}

	// Generate tokens and save refresh token
	user := &model.User{
		ID:    userID,
		Name:  registerData.Name,
		Email: registerData.Email,
	}

	jwt, refreshToken, err := s.generateAndSaveTokens(ctx, user)
	if err != nil {
		metaData["context"] = "generateAndSaveTokens"
		return nil, apperror.ErrInternalServer(err, s.Logger, metaData)
	}

	// Prepare response
	result := &common.RegisterResult{
		User:         user,
		JWT:          jwt,
		RefreshToken: refreshToken.Plaintext,
		ExpiresAt:    refreshToken.Expiry,
	}

	return result, nil
}

// AuthService handles business logic for user authentication
func (s *AccountService) AuthService(ctx context.Context, req *common.AuthRequest) (*common.AuthResult, error) {
	metaData := common.Envelop{
		"op": "service.AuthService",
	}

	// Sanitize Request
	loginData, err := validator.SanitizeLoginRequest(req)
	if err != nil {
		metaData["email"] = req.Email
		return nil, err
	}

	// Find user in database
	user, err := s.store.FindUserByEmail(ctx, loginData.Email)
	if err != nil {
		return nil, err
	}

	// Compare Password
	if !hashing.IsPasswordMatch(loginData.Password, user.PasswordHashed) {
		metaData["context"] = "Unauthorized"
		return nil, apperror.ErrUnauthorized(errors.New("invalid credentials"), s.logger, metaData)
	}

	// Generate and save tokens
	jwt, refreshToken, err := s.generateAndSaveTokens(ctx, user)
	if err != nil {
		metaData["context"] = "generateAndSaveTokens"
		return nil, apperror.ErrInternalServer(err, s.Logger, metaData)
	}

	result := &common.AuthResult{
		User:         user,
		JWT:          jwt,
		RefreshToken: refreshToken.Plaintext,
		ExpiresAt:    refreshToken.Expiry,
	}

	return result, nil
}

// LogoutService handles logout business logic
func (s *AccountService) LogoutService(ctx context.Context, refreshTokenPlaintext string) error {
	metaData := common.Envelop{
		"op": "service.LogoutService",
	}

	user, dbHash, err := s.getUserWithRefreshToken(ctx, refreshTokenPlaintext)
	if err != nil {
		return err
	}

	if err := s.tokenStore.DeleteRefreshToken(ctx, user.ID, dbHash[:]); err != nil {
		metaData["context"] = "REFRESH_TOKEN_DELETION"
		return apperror.ErrDatabaseTimeout(err, s.logger, metaData)
	}

	return nil
}

// RefreshService handles refreshRequest
func (s *AccountService) RefreshService(ctx context.Context, refreshTokenPlaintext string) (*common.AuthResult, error) {
	metaData := common.Envelop{
		"op": "service.RefreshService",
	}
	// 1. Get user and validate the incoming refresh token
	user, dbHash, err := s.getUserWithRefreshToken(ctx, refreshTokenPlaintext)
	if err != nil {
		return nil, err
	}

	// 2. Revoke/Delete the OLD refresh token (the one just used)
	// This is critical for "one-time-use" refresh tokens or token rotation.
	if err := s.tokenStore.DeleteRefreshToken(ctx, user.ID, dbHash); err != nil {
		metaData["details"] = "DELETION_FAILED_OLD_TOKEN"
		s.Logger.Errorf(fmt.Sprintf("Failed to delete used refresh token for user %d:", user.ID), err, metaData)
		return nil, apperror.ErrInternalServer(err, s.logger, metaData)
	}

	// 3. Generate and save NEW Tokens (JWT and NEW Refresh Token)
	jwt, refreshToken, err := s.generateAndSaveTokens(ctx, user)
	if err != nil {
		metaData["details"] = "TOKEN_PAIR_GENERATION_SAVE"
		s.Logger.Errorf(fmt.Sprintf("Failed to generate and save new tokens for user %d:", user.ID), err, metaData)
		return nil, apperror.ErrInternalServer(err, s.Logger, metaData)
	}

	// 4. Prepare Result for Handler
	result := &common.AuthResult{
		User:         user,
		JWT:          jwt,
		RefreshToken: refreshToken.Plaintext,
		ExpiresAt:    refreshToken.Expiry,
	}

	return result, nil
}

func (s *AccountService) AddToCollection(ctx context.Context, req *common.CollectionRequest) (*common.CollectionSuccess, error) {
	metaData := make(common.Envelop)
	// Sanitize Request
	collectionInput, err := validator.SanitizeCollectionReq(req)
	if err != nil {
		metaData["error"] = "Collection request sanitization"
		return nil, apperror.ErrBadRequest(err, s.Logger, metaData)
	}

	// Get user from context
	user, err := ctxutils.GetUser(ctx)
	if err != nil {
		metaData["warning"] = "User not found in context"
		return nil, apperror.ErrUserNotFound(err, s.logger, metaData)
	}

	// Save movie to collection
	success, err := s.store.SaveCollection(ctx, user.UserID, collectionInput.MovieID, collectionInput.Collection)
	if err != nil {
		return nil, err
	}

	var str strings.Builder
	str.WriteString("Successfully added movie: ")
	str.WriteString(strconv.Itoa(*req.MovieID))
	str.WriteString(" to the '")
	str.WriteString(*req.Collection)
	str.WriteString("' of the user: ")
	str.WriteString(strconv.Itoa(user.UserID))
	resp := &common.CollectionSuccess{
		Success: success,
		Message: str.String(),
	}
	s.logger.Info(str.String())
	return resp, nil
}

func (s *AccountService) RemoveFromCollection(ctx context.Context, req *common.CollectionRequest) (*common.CollectionSuccess, error) {
	metaData := make(common.Envelop)
	// Sanitize Request
	collectionInput, err := validator.SanitizeCollectionReq(req)
	if err != nil {
		metaData["error"] = "Collection request sanitization"
		return nil, apperror.ErrBadRequest(err, s.Logger, metaData)
	}

	// Get user from context
	user, err := ctxutils.GetUser(ctx)
	if err != nil {
		metaData["warning"] = "User not found in context"
		return nil, apperror.ErrUserNotFound(err, s.logger, metaData)
	}

	// Save movie to collection
	success, err := s.store.RemoveMovieFromCollection(ctx, user.UserID, collectionInput.MovieID, collectionInput.Collection)
	if err != nil {
		return nil, err
	}

	var str strings.Builder
	str.WriteString("Successfully removed movie: ")
	str.WriteString(strconv.Itoa(*req.MovieID))
	str.WriteString(" from the '")
	str.WriteString(*req.Collection)
	str.WriteString("' of user: ")
	str.WriteString(strconv.Itoa(user.UserID))
	resp := &common.CollectionSuccess{
		Success: success,
		Message: str.String(),
	}
	s.logger.Info(str.String())
	return resp, nil
}

func (s *AccountService) UserUpdateService(ctx context.Context, userID int, req *common.UserUpdateRequest) error {
	op := "AccountService.UserUpdateService"
	metaData := common.Envelop{
		"user_id": userID,
		"op":      op,
	}

	email, err := validator.ValidateEmail(req.Email)
	if err != nil {
		return err
	}

	name, err := validator.ValidateName(req.Name)
	if err != nil {
		return err
	}

	currentUser, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if email != "" {
		currentUser.Email = email
	}

	if name != "" {
		currentUser.Name = name
	}

	if req.ProfilePictureURL != nil {
		reqProfilePictureUrl := strings.TrimSpace(*req.ProfilePictureURL)

		if reqProfilePictureUrl != "" {
			currentUser.ProfilePictureURL = &reqProfilePictureUrl
		}
	}

	// Clean up old profile picture file if it exists and is local
	profilePictureDirPath := s.config.ProfilePicturePath
	profilePictureBase := s.config.ProfilePictureBase
	if currentUser.ProfilePictureURL != nil && *currentUser.ProfilePictureURL != "" && strings.HasPrefix(*currentUser.ProfilePictureURL, profilePictureBase) {
		oldFileName := strings.TrimPrefix(*currentUser.ProfilePictureURL, profilePictureBase)
		oldFilePath := filepath.Join(profilePictureDirPath, oldFileName)
		if _, err := os.Stat(oldFilePath); err == nil { // Check if old file exists
			if removeErr := os.Remove(oldFilePath); removeErr != nil {
				s.logger.Warn("Warning: Could not delete old profile picture '%s': %v\n", oldFilePath, removeErr, "meta", metaData)
			}
		}
	}

	err = s.store.UpdateUser(ctx, currentUser)
	if err != nil {
		return err
	}

	return nil
}

func (s *AccountService) UploadProfilePictureService(ctx context.Context, userID int, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	metaData := common.Envelop{
		"op":        "AccountService.UploadProfilePictureService",
		"file_name": fileHeader.Filename,
		"file_size": fileHeader.Size,
		"file_ext":  fileExtension,
		"user_id":   userID,
	}

	// STEP 1: VALIDATE FILE SIZE
	if fileHeader.Size > int64(common.MaxFileSize) {
		return "", apperror.ErrFileTooLarge(errors.New(apperror.ErrFileTooLargeMsg), s.logger, metaData)
	}

	// STEP 1 : VALIDATE FILE FORMAT
	if !imgutils.IsValidFileExtension(fileExtension) {
		return "", apperror.ErrUnsupportedFileType(errors.New(apperror.ErrUnsupportedFileTypeMsg), s.logger, metaData)
	}

	// STEP 2 : VALIDATE FILE MIME TYPE
	if valid, err := imgutils.IsValidImgMIMEType(file, s.logger); err != nil || !valid {
		return "", apperror.ErrUnsupportedFileType(errors.New(apperror.ErrUnsupportedFileTypeMsg), s.logger, metaData)
	}

	profilePictureDirPath := s.config.ProfilePicturePath

	// STEP 3: ENSURE DIRECTORY PATH EXISTS
	err := utils.EnsureDirExists(profilePictureDirPath)
	if err != nil {
		return "", apperror.ErrDirectoryNotFound(err, s.logger, metaData)
	}

	// STEP 3: GENERATE UNIQUE FILENAME & FILEPATH
	uniqueFilename := fmt.Sprintf("%d_%s%s", userID, time.Now().Format("20060102150405"), fileExtension)
	filePath := filepath.Join(profilePictureDirPath, uniqueFilename)

	metaData["file_path"] = filePath

	// STEP 4: RESET FILE POINTER TO START BEFORE DECODING
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", apperror.ErrFileCreateFailed(err, s.logger, metaData)
	}

	// STEP 5: OPEN THE IMAGE USING IMAGING PACKAGE
	img, err := imaging.Decode(file)
	if err != nil {
		return "", apperror.ErrFileCreateFailed(err, s.logger, metaData)
	}

	// STEP 6: RESIZE IMAGE - KEEP ASPECT RATIO WITH A MAX WIDTH OF 800px (for profile picture)
	resizedImg := imaging.Resize(img, 400, 0, imaging.Lanczos)

	// STEP 7: COMPRESS THE IMAGE (JPEG FORMAT)
	// Save the resized image to the destination file with 85% quality
	dest, err := os.Create(filePath)
	if err != nil {
		return "", apperror.ErrFileCreateFailed(err, s.logger, metaData)
	}

	defer dest.Close()

	// STEP 8: SAVE COMPRESSED IMAGE TO DISK
	err = imaging.Encode(dest, resizedImg, imaging.JPEG, imaging.JPEGQuality(85))
	if err != nil {
		return "", apperror.ErrFileCreateFailed(err, s.logger, metaData)
	}

	// STEP 9: CONSTRUCT PUBLIC URL FOR FILE
	profileImgUrl := fmt.Sprintf("%s%s", s.config.ProfilePictureBase, uniqueFilename)

	// STEP 10: SAVE PROFILE PICTURE URL
	err = s.store.SaveProfilePictureUrl(ctx, userID, profileImgUrl)
	if err != nil {
		return "", err
	}

	return profileImgUrl, nil
}

func (s *AccountService) AccountDetailsService(ctx context.Context, email string) (*model.User, error) {
	metaData := common.Envelop{
		"op":    "service.AccountDetailsService",
		"email": email,
	}

	// FIND USER BY EMAIL
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// GET FAVORITE LIST
	favoriteCollection, err := s.store.GetMovieList(ctx, "favorites", user.ID)
	if err != nil {
		return nil, err
	}

	if len(favoriteCollection) == 0 {
		s.logger.Info("User has no item in favorite collection", "meta", metaData)
	}
	user.Favorites = favoriteCollection

	// GET WATCHLIST LIST
	watchlistCollection, err := s.store.GetMovieList(ctx, "watchlist", user.ID)
	if err != nil {
		return nil, err
	}

	if len(watchlistCollection) == 0 {
		s.logger.Info("User has no item in watchlist collection", "meta", metaData)
	}
	user.Watchlist = watchlistCollection

	return user, nil
}

// generateAndSaveToken helper method generates access_token and refresh_token and saves refresh_token
func (s *AccountService) generateAndSaveTokens(ctx context.Context, user *model.User) (string, *tokens.Token, error) {
	metaData := common.Envelop{
		"userID": user.ID,
	}
	// Generate tokens (JWT(AccessToken) and RefreshToken)
	jwt, refreshToken, err := s.tokens.GenerateTokenPair(user)
	if err != nil {
		return "", nil, apperror.ErrInternalServer(err, s.Logger, metaData)
	}

	// Save refresh token hash in store
	refreshTokenSaveErr := s.tokenStore.SaveRefreshToken(ctx, refreshToken)
	if refreshTokenSaveErr != nil {
		return "", nil, apperror.ErrInternalServer(refreshTokenSaveErr, s.Logger, metaData)
	}

	return jwt, refreshToken, nil
}

func (s *AccountService) getUserWithRefreshToken(ctx context.Context, plainTextToken string) (*model.User, []byte, error) { // Accepts string
	metaData := common.Envelop{
		"op": "service.getUserWithRefreshToken",
	}

	// Sanitize (basic check, more complex in validator)
	if plainTextToken == "" {
		metaData["details"] = "empty_refresh_token_provided_to_service"
		return nil, nil, apperror.ErrBadRequest(nil, s.Logger, metaData)
	}

	inputTokenHash := sha256.Sum256([]byte(strings.TrimSpace(plainTextToken)))

	storedToken, err := s.tokenStore.GetTokenDetailsByHash(ctx, inputTokenHash[:])
	if err != nil {
		s.logger.Error("failed to get token details (db lookup by hash)", err, metaData)
		return nil, nil, err
	}

	metaData["user_id"] = storedToken.UserID

	if err := s.tokens.VerifyRefreshToken(plainTextToken, storedToken.Hash); err != nil {
		metaData["details"] = "invalid_refresh_token_hash_mismatch"
		s.Logger.Error("Provided refresh token plaintext does not match stored hash", err, metaData)
		return nil, nil, apperror.ErrUnauthorized(err, s.Logger, metaData)
	}

	if time.Now().After(storedToken.Expiry) {
		s.Logger.Info("Expired refresh token found during logout attempt", metaData)
		if err = s.tokenStore.DeleteRefreshToken(ctx, storedToken.UserID, storedToken.Hash); err != nil {
			metaData["warning"] = true
			s.logger.Error("Failed to delete expired refresh token during logout cleanup", err, metaData)
		}
		metaData["details"] = "refresh_token_expired"
		return nil, nil, apperror.ErrTokenExpired(nil, s.Logger, metaData)
	}

	user, err := s.store.FindUserByID(ctx, storedToken.UserID)
	if err != nil {
		s.Logger.Errorf(fmt.Sprintf("Failed to find user by ID %d for logout: %+v", storedToken.UserID, err), err, metaData)
		return nil, nil, err
	}

	if user == nil {
		metaData["details"] = "user_not_found_for_refresh_token_after_db_lookup"
		return nil, nil, apperror.ErrUnauthorized(nil, s.Logger, metaData)
	}

	return user, storedToken.Hash, nil
}

func (s *AccountService) DeleteTokenService(ctx context.Context, id int) error {
	return s.tokenStore.DeleteAllTokensForUser(ctx, id)
}

// RequestEmailVerification initiates the verification email flow.
func (s *AccountService) RequestEmailVerification(ctx context.Context, email string) error {
	op := "service.RequestEmailVerification"
	meta := common.Envelop{"email": email, "op": op}

	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Info("Verification requested for unknown email (slient fial)", meta)
	}

	if user.IsVerified {
		s.logger.Info("User is already verified, skipping email send", meta)
		return nil
	}

	verificationTTL := 24 * time.Hour
	token, err := s.generateAndSaveAuthToken(ctx, user, tokens.EmailVerificationScope, verificationTTL)
	if err != nil {
		s.logger.Error("Failed to generate/save verification token", err, meta)
		return apperror.ErrInternalServer(err, s.logger, meta)
	}

	// Finally, sending email
	if err := s.emailSender.SendVerificationEmail(user.Email, user.Name, token.Plaintext); err != nil {
		s.Logger.Warn("Failed to send verification email. Token remains in DB.", err, meta)
		return nil
	}

	return nil //🚫 SECURITY: SILENT FIAL
}

// RequestPasswordReset handles finding the user, generating the reset token,
// saving the token hash, and sending the reset email.
func (s *AccountService) RequestPasswordReset(ctx context.Context, email string) error {
	op := "service.RequestPasswordReset"
	meta := common.Envelop{"email": email, "op": op}

	// 1. Find User
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		// IMPORTANT: Security by obscurity. Suppress the error (like ErrUserNotFound)
		// to prevent an attacker from enumerating registered email addresses.
		s.Logger.Info("Password reset requested for unknown email (silent fail)", meta)
		return nil // Always return nil (success) for the public endpoint
	}

	// 2. Generate and Save Single-Use Token
	resetTTL := 1 * time.Hour

	// Creating and saving the single-use token.
	token, err := s.generateAndSaveAuthToken(
		ctx,
		user,
		tokens.PasswordResetScope, // Using the correct scope constant
		resetTTL,
	)
	if err != nil {
		// Return an internal server error if token generation/saving fails.
		s.Logger.Error("Failed to generate and save password reset token", err, meta)
		return apperror.ErrInternalServer(err, s.Logger, meta)
	}

	// 3. Send Email
	// If the email sending fails, we log it but usually continue to return nil
	// to the external caller for security reasons (don't reveal the exact failure point).
	if err := s.emailSender.SendPasswordResetEmail(user.Email, user.Name, token.Plaintext); err != nil {
		s.Logger.Warn("Failed to send password reset email. Token remains in DB.", err, meta)
		// We can optionally delete the token here if email delivery failure is critical,
		// but often we rely on the 1-hour TTL for cleanup.
	}

	s.Logger.Info("Password reset initiated and email dispatched", meta)
	return nil
}

// VerifyEmail confirms the email using the generic processor.
func (s *AccountService) VerifyEmail(ctx context.Context, plainTextToken string) error {
	op := "AccountService.VerifyEmail"

	// Mark user verified in db
	verifyAction := func(ctx context.Context, userID int) error {
		return s.store.MarkUserAsVerified(ctx, userID)
	}

	return s.processAuthToken(ctx, plainTextToken, tokens.EmailVerificationScope, op, verifyAction)
}

// ConfirmPasswordReset confirms the reset using the generic processor.
func (s *AccountService) ConfirmPasswordReset(ctx context.Context, plainTextToken, newPassword string) error {
	op := "service.ConfirmPasswordReset"
	meta := common.Envelop{"op": op}

	// 1. Validate New Password
	if err := validator.ValidatePassword(newPassword); err != nil {
		return apperror.ErrBadRequest(err, s.logger, meta)
	}

	// 2. Hash New Password
	newPass := strings.TrimSpace(newPassword)
	newHashedPassword, err := hashing.SetHash(newPass)
	if err != nil {
		return apperror.ErrInternalServer(err, s.Logger, meta)
	}

	resetAction := func(ctx context.Context, userID int) error {
		return s.store.UpdatePassword(ctx, userID, newHashedPassword)
	}

	return s.processAuthToken(ctx, plainTextToken, tokens.PasswordResetScope, op, resetAction)
}

// processAuthToken performs the secure, single-use token lifecycle:
// 1. Hashes the plaintext token.
// 2. Looks up the token hash in the store (by hash and scope).
// 3. Checks for expiration.
// 4. Executes a custom user action (business logic).
// 5. Deletes the token (consumption).
func (s *AccountService) processAuthToken(ctx context.Context, plainTextToken, scope, op string, action func(ctx context.Context, userID int) error) error {
	meta := common.Envelop{"op": op, "scope": scope}
	// 1. Hash the incoming plaintext token
	tokenHash := sha256.Sum256([]byte(strings.TrimSpace(plainTextToken)))

	// 2. Look up the token by hash and scope
	storedToken, err := s.tokenStore.GetAuthTokenByHash(ctx, tokenHash[:], scope)
	if err != nil {
		s.Logger.Error("Token lookup failed", err, meta)
		return apperror.ErrTokenMalformed(err, s.Logger, meta)
	}

	meta["user_id"] = storedToken.UserID

	// 3. Check for expiration: Clean up expired token
	if time.Now().After(storedToken.Expiry) {
		s.tokenStore.DeleteAuthTokenByHash(ctx, tokenHash[:])
		s.Logger.Info("Token expired during processing", meta)
		return apperror.ErrTokenExpired(nil, s.Logger, meta)
	}

	// 4. Execute the custom business logic (e.g., Verify Email, Update Password)
	if err := action(ctx, storedToken.UserID); err != nil {
		s.Logger.Error("Custom token action failed", err, meta)
		return err
	}

	// 5. CRITICAL: Consume/Delete the token
	if err := s.tokenStore.DeleteAuthTokenByHash(ctx, tokenHash[:]); err != nil {
		s.Logger.Warn("Warning: Failed to delete used token after successful action", err, meta)
		return nil // Log the warning, but return nil since the user action succeeded.
	}

	s.Logger.Info("Token successfully processed and consumed", meta)

	return nil
}

// generateAndSaveAuthToken creates a single-use token and saves its hash.
func (s *AccountService) generateAndSaveAuthToken(
	ctx context.Context,
	user *model.User,
	scope string,
	ttl time.Duration,
) (*tokens.Token, error) {
	op := "service.generateAndSaveAuthToken"
	meta := common.Envelop{"user_id": user.ID, "scope": scope}

	token, err := s.tokens.CreateRefreshToken(user.ID, ttl, scope)
	if err != nil {
		return nil, apperror.ErrInternalServer(err, s.Logger, meta)
	}

	if err := s.tokenStore.SaveAuthToken(ctx, token); err != nil {
		return nil, apperror.ErrDatabaseOpFailed(apperror.CodeDatabaseError, "Failed to save auth token", op, err, s.Logger, meta)
	}

	return token, nil
}

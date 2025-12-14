package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"multipass/internal/model"
	"multipass/internal/service"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/cookieutils"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/response"
	"multipass/pkg/utils"
)

type AccountHandler struct {
	BaseHandler
	service service.UserAccountService
	// emailService service.EmailService
}

func NewAccountHandler(service service.UserAccountService,
	// emailService service.EmailService,
	logger logging.Logger, responder response.Writer,
) *AccountHandler {
	return &AccountHandler{
		service: service,
		// emailService: emailService,
		BaseHandler: BaseHandler{
			Logger:       logger,
			Responder:    responder,
			ErrorHandler: apperror.NewBaseErrorHandler(logger, responder),
		},
	}
}

// SignUp handles user registration
func (h *AccountHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metaData := common.Envelop{
		"op":     "AccountHandler.SignUp",
		"method": r.Method,
		"path":   r.URL.Path,
	}

	// Decode, Sanitize, and Validate
	req, err := utils.DecodeRequest[common.RegisterRequest](w, r, "User registration")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "user registration request") {
			return
		}
	}

	result, err := h.service.RegisterUser(ctx, req)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "user registration") {
			return
		}
	}

	resp := &common.AuthResponse{
		Success: true,
		Message: "Successfully registered user",
		JWT:     result.JWT,
	}

	metaData["email"] = result.User.Email

	cookieutils.SetCookie(w, h.Logger, "refresh_token", result.RefreshToken, "/", 7*time.Now().Day(), result.ExpiresAt)

	if err := h.Responder.WriteJSON(w, http.StatusCreated, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}
	h.Logger.Info("successfully processed user registration request", "meta", metaData)
}

// Login handles user login authentication route
func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.Login",
	}

	// Decode Authentication Request
	req, err := utils.DecodeRequest[common.AuthRequest](w, r, "User authentication")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "from user authentication request") {
			return
		}
	}

	result, err := h.service.AuthService(ctx, req)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "user registration") {
			return
		}
	}

	// Prepare response
	resp := &common.AuthResponse{
		Success: true,
		Message: "Successfully authenticated user",
		JWT:     result.JWT,
		User:    result.User,
	}

	metaData["email"] = result.User.Email

	cookieutils.SetCookie(w, h.Logger, "refresh_token", result.RefreshToken, "/", 7*time.Now().Day(), result.ExpiresAt)

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user authorization request", "meta", metaData)
}

// Refresh handles user token refresh request
func (h *AccountHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metadata := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.Refresh",
	}

	// 1. Extract Refresh Token from HttpOnly Cookie
	cookie, err := cookieutils.GetCookie(r, h.Logger, "refresh_token")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r,
			apperror.ErrUnauthorized(err, h.Logger, metadata), "refresh_token_cookie") {
			return
		}
		return
	}

	// 2. Call the Refresh Service
	refreshTokenPlaintext := cookie.Value
	result, err := h.service.RefreshService(ctx, refreshTokenPlaintext)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "refresh_service_execution") {
			return
		}
		return
	}

	metadata["email"] = result.User.Email
	resp := &common.RefreshResponse{
		JWT: result.JWT,
	}

	// 4. Set NEW HttpOnly Refresh Token Cookie
	cookieutils.SetCookie(w, h.Logger, "refresh_token", result.RefreshToken, "/", 7*time.Now().Day(), result.ExpiresAt)

	// 5. Send Success Response (new JWT in body)
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data": resp,
	}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metadata), "response writer") {
			return
		}
	}
	// 6. Final Handler Logging
	h.Logger.Info("successfully processed user refresh request", metadata)
}

// Logout handles user logout request
func (h *AccountHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := "AccountHandler.Logout"
	metaData := common.Envelop{
		"op":     op,
		"method": r.Method,
		"path":   r.URL.Path,
	}

	cookie, err := cookieutils.GetCookie(r, h.Logger, "refresh_token")
	if err != nil {
		h.Logger.Info("Logout attempt with no refresh_token cookie found. Assuming already logged out.", metaData)

		if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": common.AuthResponse{Success: true, Message: "Successfully logged out."}}); err != nil {
			h.Logger.Errorf("Error writing JSON success response for logout (no cookie): %+w", err)
		}
		return
	}

	refreshTokenPlaintext := cookie.Value
	if err := h.service.LogoutService(ctx, refreshTokenPlaintext); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "logout_service_execution") {
			return
		}
		return
	}

	cookieutils.SetCookie(w, h.Logger, "refresh_token", "", "/", time.Now().Second()*5, time.Now().Add(-24*time.Hour))

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": common.AuthResponse{
		Success: true,
		Message: "You have been successfully logged out",
	}}); err != nil {

		h.Logger.Errorf("Error writing JSON success response for logout: %+w", err)
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "final logout response write")
	}

	h.Logger.Info("User logout request fully processed", metaData)
}

// HandleSaveToCollection saves movie to favorite/watchlist collections
func (h *AccountHandler) HandleSaveToCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"status": "error",
		"method": r.Method,
		"path":   r.URL.Path,
	}
	req, err := utils.DecodeRequest[common.CollectionRequest](w, r, "SaveToCollection Request")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "SaveToCollection request") {
			return
		}
	}

	resp, err := h.service.AddToCollection(ctx, req)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "add movie to user's collection") {
			return
		}
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data": resp,
	}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}
}

// HandleRemoveFromCollection removes movie from favorite/watchlist collections
func (h *AccountHandler) HandleRemoveMovieFromCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"status": "error",
		"method": r.Method,
		"path":   r.URL.Path,
	}
	req, err := utils.DecodeRequest[common.CollectionRequest](w, r, "Remove Movie From Collection Request")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "RemoveMovieFromCollection request") {
			return
		}
	}

	resp, err := h.service.RemoveFromCollection(ctx, req)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "remove movie from user collection") {
			return
		}
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data": resp,
	}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}
}

// HandleGetFavorites retrieves user's favorite collection
func (h *AccountHandler) HandleGetFavorites(w http.ResponseWriter, r *http.Request) {
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.HandleGetUserProfile",
	}

	accountDetails, err := h.handleUserDetails(r, metaData)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "GetFavorites_AccountDetailsService") {
			return
		}
	}

	resp := &common.MoviesResponse{
		Success: true,
		Movies:  accountDetails.Favorites,
		Count:   len(accountDetails.Favorites),
	}
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data": resp,
	}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}
}

// HandleGetFavorites retrieves user's watchlist collection
func (h *AccountHandler) HandleGetWatchlist(w http.ResponseWriter, r *http.Request) {
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.HandleGetUserProfile",
	}

	accountDetails, err := h.handleUserDetails(r, metaData)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "GetWatchlist_AccountDetailsService") {
			return
		}
	}

	resp := &common.MoviesResponse{
		Success: true,
		Movies:  accountDetails.Watchlist,
		Count:   len(accountDetails.Watchlist),
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"data": resp,
	}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, nil), "response writer") {
			return
		}
	}
}

func (h *AccountHandler) HandleUserUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.HandleUserUpdate",
	}

	userID, err := utils.GetParamID(r)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "param_user_id") {
			return
		}
		return
	}
	metaData["user_id"] = userID

	req, err := utils.DecodeRequest[common.UserUpdateRequest](w, r, "user_update")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "from user authentication request") {
			return
		}
		return
	}

	err = h.service.UserUpdateService(ctx, userID, req)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "user_update_service") {
			return
		}
		return
	}

	resp := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: true,
		Message: "Profile updated successfully",
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user update request", "meta", metaData)
}

func (h *AccountHandler) HandleProfilePictureUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := "AccountHandler.HandleProfilePictureUpload"
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     op,
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// PARSE MULTIPART FORM
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		appErr := apperror.ErrFileTooLarge(errors.New(apperror.ErrFileTooLargeMsg), h.Logger, metaData)
		if h.ErrorHandler.HandleAppError(w, r, appErr, "ParseMultipartForm") {
			return
		}
		return
	}

	user, err := ctxutils.GetUser(ctx)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrCtxUserMissing(err, h.Logger, metaData), "ctxutils.GetUser") {
			return
		}
		return
	}

	file, fileHeader, err := r.FormFile("profilePicture")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrBadRequest(err, h.Logger, metaData), "AccountHandler.FormFile") {
			return
		}
		return
	}

	defer file.Close()

	profileImgURL, err := h.service.UploadProfilePictureService(ctx, user.UserID, file, fileHeader)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrFileUploadFailed(err, h.Logger, metaData), op) {
			return
		}
		return
	}

	resp := &common.ProfilePictureUploadResponse{
		Success:           true,
		ProfilePictureURL: profileImgURL,
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully uploaded profile picture", "meta", metaData)
}

func (h *AccountHandler) HandleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "AccountHandler.HandleGetUserProfile",
	}

	profileDetails, err := h.handleUserDetails(r, metaData)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "param_user_id") {
			return
		}
		return
	}

	resp := &common.UserProfileResponse{
		Success: true,
		Message: "Successfully retrieved user profile",
		User:    profileDetails,
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user profile request", "meta", metaData)
}

func (h *AccountHandler) handleUserDetails(r *http.Request, meta common.Envelop) (*model.User, error) {
	ctx := r.Context()
	meta["op"] = "AccountHandler.handleUserDetails"

	user, err := ctxutils.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found in context %w", err)
	}

	if user == nil {
		return nil, apperror.ErrUserNotFound(errors.New("context user is nil"), h.Logger, meta)
	}

	accountDetails, err := h.service.AccountDetailsService(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	return accountDetails, nil
}

// RequestPasswordReset handles the request to send a password reset link to a user's email.
func (h *AccountHandler) HandlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{"op": "AccountHandler.RequestPasswordReset"}

	// 1. Decode request
	req, err := utils.DecodeRequest[struct{ Email string }](w, r, "Password reset request")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "password reset request") {
			return
		}
	}

	// 2. Business Logic
	h.service.RequestPasswordReset(ctx, req.Email)
}

// -----------------------------------------------------------
// EMAIL VERIFICATION
// -----------------------------------------------------------

// VerifyEmail handles the token-based email verification process.
// Route: GET /v1/account/verify-email?token={token}
func (h *AccountHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := "AccountHandler.emailVefifyHandler"
	metaData := common.Envelop{"op": op}

	// 1. Get token from URL query
	token := r.URL.Query().Get("token")
	if token == "" {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrTokenMissing(errors.New("missing verification token"), h.Logger, metaData), "missing_token")
	}

	// 2. Process Token
	if err := h.service.VerifyEmail(ctx, token); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "verification_service_failed") {
			return
		}
		return
	}

	// 3. Success Response
	resp := common.GenericResponse{
		Success: true,
		Message: "Email successfully verified. You can now log in.",
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		return
	}
	h.Logger.Info("Email verification successfully processed", metaData)
}

// -----------------------------------------------------------
// PASSWORD RESET CONFIRMATION
// -----------------------------------------------------------

// ConfirmPasswordReset handles the final step of the password reset flow.
// It requires the token from the URL and the new password in the body.
// Route: POST /v1/account/confirm-reset
func (h *AccountHandler) HandleConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"op":     "AccountHandler.ConfirmPasswordReset",
		"method": r.Method,
		"path":   r.URL.Path,
	}

	// 1. Decode request
	req, err := utils.DecodeRequest[common.PasswordResetRequest](w, r, "Password_Reset_Request")
	if err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "reset confirm request")
		return
	}

	// 2: Delegating Password Reset Request to Account Service
	if err := h.service.ConfirmPasswordReset(ctx, req.Token, req.NewPassword); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "password_reset_service_failed") {
			return
		}
		return
	}

	// 3: Success Response
	resp := common.GenericResponse{
		Success: true,
		Message: "Password successfully reset. Please log in with your new password.",
	}

	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "final response write")
	}
	h.Logger.Info("Password reset successfully confirmed and consumed", metaData)
}

// src/internals/api/account_handler.go (Additions)

// -----------------------------------------------------------
// OTP (ONE-TIME PASSWORD) HANDLERS
// -----------------------------------------------------------

// RequestOTP handles the request to send an OTP code to a user's email.
// Route: POST /v1/account/request-otp
// func (h *AccountHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	op := "AccountHandler.RequestOTP"
// 	metaData := common.Envelop{"op": op}

// 	// 1. Decode request
// 	req, err := utils.DecodeRequest[common.SendOTPRequest](w, r, "send_otp_request")
// 	if err != nil {
// 		h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "otp request decode")
// 		return
// 	}

// 	// 2. Business Logic (Silent Fail)
// 	h.service.RequestOTP(ctx, req.Email)

// 	// 3. Success Response (Security: Always send success to prevent email enumeration)
// 	resp := common.GenericResponse{
// 		Success: true,
// 		Message: "If the account exists, an OTP code has been sent to the email address.",
// 	}

// 	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
// 		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "final response write")
// 	}
// 	h.Logger.Info("OTP request successfully processed (silent fail applied)", metaData)
// }

// VerifyOTP validates the OTP code provided by the user.
// Route: POST /v1/account/verify-otp
// func (h *AccountHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	metaData := common.Envelop{"op": "AccountHandler.VerifyOTP"}

// 	// 1. Decode request
// 	req, err := utils.DecodeRequest[common.OTPRequest](w, r, "OTP_Request")
// 	if err != nil {
// 		h.ErrorHandler.HandleAppError(w, r, apperror.ErrJSONDecodeFailed(err, h.Logger, metaData), "otp verification decode")
// 		return
// 	}
// 	metaData["email"] = req.Email

// 	// 2. Call Service Layer to Verify
// 	if err := h.service.VerifyOTP(ctx, req.Email, req.OTP); err != nil {
// 		// This will catch ErrUnauthorized, ErrTokenExpired, etc.
// 		if h.ErrorHandler.HandleAppError(w, r, err, "otp verification failed") {
// 			return
// 		}
// 		return
// 	}

// 	// 3. Success Response
// 	resp := common.GenericResponse{
// 		Success: true,
// 		Message: "OTP successfully verified.",
// 	}

// 	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
// 		h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "final response write")
// 	}
// 	h.Logger.Info("OTP successfully verified and consumed", metaData)
// }

// user, err := ctxutils.GetUser(ctx)
// if err != nil {
// 	appErr := apperror.ErrUserNotFound(err, h.Logger, metaData)
// 	metaData["error"] = appErr
// 	h.Logger.Warn("User not found in context", "meta", metaData)
// 	if h.ErrorHandler.HandleAppError(w, r, appErr, "param_user_id") {
// 		return
// 	}
// 	return
// }

// // # sendErr helper sends WriteJSON error message
// func (h *AccountHandler) sendErr(w http.ResponseWriter, r *http.Request, err error) {
// 	errMeta := common.Envelop{
// 		"method":     r.Method,
// 		"path":       r.URL.Path,
// 		"statusCode": http.StatusInternalServerError,
// 	}
// 	resErr := apperror.ErrFailedJSONWriter(err, h.Logger, errMeta)
// 	h.Logger.Error(apperror.ErrFailedJSONResWrite, resErr, errMeta)
// 	resErr.WriteJSONError(w, r, h.Responder)
// }

// // # handleError a helper to centralize error handling
// func (h *AccountHandler) handleError(w http.ResponseWriter, r *http.Request, err error, code int, context string) bool {
// 	if err == nil {
// 		return false
// 	}

// 	// Base metadata for logging and responses
// 	errMeta := common.Envelop{
// 		"method": r.Method,
// 		"path":   r.URL.Path,
// 	}

// 	// Check if the error is an AppError, if so, handle it specifically
// 	var appErr *apperror.AppError

// 	if errors.As(err, &appErr) {
// 		appErr.Code = code
// 		appErr.Metadata = errMeta
// 		appErr.Message = apperror.FormatErrMsg(context, appErr.Err)
// 		appErr.Logger = h.Logger
// 		appErr.LogError()
// 		appErr.WriteJSONError(w, r, h.Responder)

// 		return true
// 	}

// func (h *AccountHandler) sendDecodeErr(w http.ResponseWriter, r *http.Request, err error, reqTypeMsg string) {
// 	h.writeAppError(w, r, apperror.ErrBadRequest(err, h.Logger, common.Envelop{
// 		"error":   fmt.Sprintf("%s %s", apperror.ErrJSONDecodeFailedMsg, reqTypeMsg),
// 		"details": err.Error(),
// 	}))
// }

// // helper function to handle errors with custom app error wrapping
// If it's not an AppError, wrap it as one for consistency
// 	wrappedErr := apperror.NewAppError(
// 		apperror.CodeInternal,
// 		"Unknown error occurred",
// 		context,
// 		err,
// 		h.Logger,
// 		errMeta,
// 	)

// 	// Log the wrapped error
// 	wrappedErr.Metadata["error"] = apperror.FormatErrMsg(context, wrappedErr)
// 	wrappedErr.LogError()
// 	wrappedErr.WriteJSONError(w, r, h.Responder)

// 	return true
// }

// func (h *AccountHandler) writeAppError(w http.ResponseWriter, r *http.Request, appErr *apperror.AppError) {
// 	if appErr == nil {
// 		// fallback in case of programming error
// 		appErr = apperror.NewAppError(
// 			apperror.CodeInternal,
// 			"Unknown application error",
// 			"writeAppError",
// 			nil,
// 			h.Logger,
// 			common.Envelop{
// 				"status": "error",
// 				"path":   r.URL.Path,
// 				"method": r.Method,
// 			},
// 		)
// 	}

// 	if appErr.Logger != nil {
// 		appErr.LogError()
// 	}

// 	appErr.WriteJSONError(w, r, h.Responder)
// }

// func (h *AccountHandler) SignUp(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	errMeta := common.Envelop{
// 		"status": "error",
// 		"method": r.Method,
// 		"path":   r.URL.Path,
// 	}

// 	// 1: Decode request 🏁
// 	req, err := utils.DecodeRequest[common.RegisterRequest](w, r, "User registration")
// 	if err != nil {
// 		h.sendDecodeErr(w, r, err, "from user registration request")
// 		return
// 	}

// 	// 2: Sanitize
// 	var registerData common.RegisterInput
// 	registerData, err = validator.SanitizeRegisterRequest(req)
// 	if err != nil {
// 		appErr := apperror.ErrBadRequest(err, h.Logger, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	// 3: Validate
// 	if err := validator.ValidateRegisterRequest(registerData); err != nil {
// 		errMeta["error"] = apperror.ErrUserRegistrationValidationFailedMsg
// 		errMeta["details"] = err.Error()
// 		appErr := apperror.ErrRegistrationValidation(err, h.Logger, errMeta)
// 		h.Logger.Errorf("failed user register validation: %+w", appErr, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	// 4: Check email uniqueness
// 	exists, err := h.accountStore.CheckEmailExists(ctx, registerData.Email)
// 	if err != nil {
// 		errMeta["error"] = apperror.ErrQueryFailedMsg
// 		appErr := apperror.ErrInternalServer(err, h.Logger, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	if exists {
// 		h.Logger.Errorf("User already exists with email: %s", req.Email)
// 		appErr := apperror.ErrEmailAlreadyRegistered(nil, h.Logger, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	// 5: Hash password
// 	passwordHash, err := hashing.SetHash(registerData.Password)
// 	if err != nil {
// 		errMeta["error"] = apperror.ErrHashingFailedMsg + " password"
// 		appErr := apperror.ErrInternalServer(err, h.Logger, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	// 6: Register user
// 	userID, err := h.accountStore.CreateUser(ctx, registerData.Name, registerData.Email, passwordHash)
// 	if h.ErrorHandler.Handle(w, r, err, apperror.ErrRegistrationFailedMsg) {
// 		return
// 	}

// 	// 7: Generate Tokens (JWT) + Refresh Token
// 	user := &models.User{
// 		ID:    userID,
// 		Name:  registerData.Name,
// 		Email: registerData.Email,
// 	}

// 	accessToken, refreshToken, err := h.tokens.GenerateTokenPair(user)
// 	if err != nil {
// 		h.writeAppError(w, r, apperror.ErrGenerateTokenFailed(err, h.Logger, errMeta))
// 		return
// 	}

// 	// 8: Store hashed refresh token
// 	refreshTokenSaveErr := h.tokenStore.SaveRefreshToken(ctx, refreshToken)
// 	if h.ErrorHandler.Handle(w, r, refreshTokenSaveErr, "tokenStore.SaveRefreshToken") {
// 		return
// 	}

// 	// 9: Respond with success
// 	resp := RegisterResponse{
// 		Success:      true,
// 		Message:      "Successfully registered user",
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken.Plaintext,
// 		ExpiresAt:    refreshToken.Expiry,
// 	}
// 	if err := h.Responder.WriteJSON(w, http.StatusCreated, common.Envelop{"data": resp}); err != nil {
// 		h.sendErr(w, r, err)
// 		return
// 	}

// 	h.Logger.Info("user successfully registered", "email", registerData.Email)
// }

// VALIDATION
// noLint:unused

// var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`) // allow letters, numbers, and underscores

//		r.logger.Error("Registration validation failed: missing required fields", nil)
//		return 0, apperror.ErrRegistrationFailedMsg
//	}
//
// if name == "" {
// func trim(s string) string {
// 	return strings.TrimSpace(s)
// }

// email := strings.TrimSpace(req.Email)
// if email == "" {
// 	return errors.New("email is required")
// }

// json.NewDecoder(r.Body).Decode(&req); err != nil

// func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	errMeta := common.Envelop{
// 		"method": r.Method,
// 		"path":   r.URL.Path,
// 	}
// 	// 1: Decode Authentication Request
// 	req, err := utils.DecodeRequest[common.AuthRequest](w, r, "User authentication")
// 	if err != nil {
// 		h.sendDecodeErr(w, r, err, "from user authentication request")
// 		return
// 	}

// 	// 2: Sanitize and Nil-Check Authentication Request
// 	loginData, err := validator.SanitizeLoginRequest(req)
// 	if err != nil {
// 		errMeta["details"] = err.Error()
// 		appErr := apperror.ErrBadRequest(err, h.Logger, errMeta)
// 		h.writeAppError(w, r, appErr)
// 		return
// 	}

// 	// 3: Find User by Email
// 	user, err := h.accountStore.FindUserByEmail(ctx, loginData.Email)
// 	if h.ErrorHandler.Handle(w, r, err, "FindUserByEmail") {
// 		return
// 	}

// 	if user == nil {
// 		h.writeAppError(w, r, apperror.ErrUserNotFound(err, h.Logger, errMeta))
// 		return
// 	}

// 	// 4: Compare Password with HASHED Password
// 	if ok := hashing.IsPasswordMatch(loginData.Password, user.PasswordHashed); !ok {
// 		h.writeAppError(w, r, apperror.ErrUnauthorizedAccess(err, h.Logger, errMeta))
// 		return
// 	}

// 	// 5: Generate Tokens (JWT) + Refresh Token
// 	accessToken, refreshToken, err := h.tokens.GenerateTokenPair(user)
// 	if err != nil {
// 		h.writeAppError(w, r, apperror.ErrGenerateTokenFailed(err, h.Logger, errMeta))
// 		return
// 	}

// 	// 6: Store hashed refresh token
// 	refreshTokenSaveErr := h.tokenStore.SaveRefreshToken(ctx, refreshToken)
// 	if h.ErrorHandler.Handle(w, r, refreshTokenSaveErr, "tokenStore.SaveRefreshToken") {
// 		return
// 	}

// 	// 6: Send response with Access Token + Refresh Token
// 	resp := AuthResponse{
// 		Success:      true,
// 		Message:      "Successfully authenticated login request",
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken.Plaintext,
// 		ExpiresAt:    refreshToken.Expiry,
// 	}

// 	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
// 		"data": resp,
// 	}); err != nil {
// 		h.sendErr(w, r, err)
// 		return
// 	}

// 	h.Logger.Info("user logged in", "email", user.Email)
// }

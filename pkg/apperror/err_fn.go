package apperror

import (
	"fmt"

	"multipass/pkg/common"
	"multipass/pkg/logging"
)

// ErrUnauthorized creates an error for required authentication.
func ErrUnauthorized(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrUnauthorizedMsg, "authentication_required", err, logger, metadata)
}

// ErrForbidden creates an error indicating the user lacks permission.
func ErrForbidden(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeForbidden, ErrForbiddenMsg, "authorization_failed", err, logger, metadata)
}

// ErrInvalidAuth creates an error for malformed or incorrect authentication details.
func ErrInvalidAuth(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrInvalidAuthMsg, "auth_header_or_token_invalid", err, logger, metadata)
}

// ErrMissingAuth creates an error when an authentication token is not provided.
func ErrMissingAuth(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrMissingAuthMsg, "auth_token_missing", err, logger, metadata)
}

// ErrInvalidJWT creates an error for an expired session or token.
func ErrInvalidJWT(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrUnauthorizedMsg, "token_validation", err, logger, metadata)
}

// ErrTokenExpired creates an error for an expired session or token.
func ErrTokenExpired(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrTokenExpiredMsg, "session_or_token_expired", err, logger, metadata)
}

// ErrTokenMalformed creates an error for a malformed token.
func ErrTokenMalformed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrTokenMalformedMsg, "token_format_malformed", err, logger, metadata)
}

// ErrInvalidAPIKey creates an error for an invalid API key.
func ErrInvalidAPIKey(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrInvalidAPIKeyMsg, "api_key_validation", err, logger, metadata)
}

// ErrInvalidTokenSignature creates an error for an invalid token signature.
func ErrInvalidTokenSignature(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrInvalidTokenSignatureMsg, "token_signature_verification_failed", err, logger, metadata)
}

// ErrCSRFTokenMissing creates an error for a missing CSRF token.
func ErrCSRFTokenMissing(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeForbidden, ErrCSRFTokenMissingMsg, "csrf_token_missing", err, logger, metadata)
}

// ErrTokenMissing creates an error for a token that is not yet active. (Clarified context/message based on distinction from ErrMissingAuth)
func ErrTokenMissing(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	// Original message was ErrMissingAuthMsg, changed to ErrTokenNotFoundMsg for clearer distinction
	return NewAppError(CodeUnauthorized, ErrTokenNotFoundMsg, "token_not_found_or_invalid_in_system", err, logger, metadata)
}

// ErrTokenNotActive creates an error for a token that is not yet active.
func ErrTokenNotActive(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrTokenNotActiveMsg, "token_not_active_yet", err, logger, metadata)
}

// ErrCookieNotFound creates an error for invalid or missing cookiee.
func ErrCookieNotFound(err error, logger logging.Logger, metadata common.Envelop, cookieName string) *AppError {
	return NewAppError(CodeBadRequest, fmt.Sprintf("Missing cookie: %s", cookieName), "getcookie", err, logger, metadata)
}

// ErrCookieNotFound creates an error for invalid or missing cookiee.
func ErrInvalidCookie(err error, logger logging.Logger, metadata common.Envelop, cookieName string) *AppError {
	return NewAppError(CodeBadRequest, fmt.Sprintf("Missing cookie: %s", cookieName), "getcookie", err, logger, metadata)
}

// ErrCookieNotFound creates an error for invalid or missing cookiee.
func ErrUnexpectedCookie(err error, logger logging.Logger, metadata common.Envelop, cookieName string) *AppError {
	return NewAppError(CodeInternal, fmt.Sprintf("Unexpected error while extracting cookie: %s", cookieName), "getcookie", err, logger, metadata)
}

// ErrEmailSendFailed creates an error for invalid or missing cookiee.
func ErrEmailSendFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, "Unexpected error while sending email", "sendEmail", err, logger, metadata)
}

// ---
// ## User Account & Profile Errors
// These functions create errors related to user registration, account states, and profile management.
// ---

// ErrUserNotFound creates an error when a user cannot be found.
func ErrUserNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrUserNotFoundMsg, "store.user_not_found", err, logger, metadata)
}

// ErrCtxUserMissing creates an error when a user cannot be found.
func ErrCtxUserMissing(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrUserNotFoundMsg, "ctx.user_not_found", err, logger, metadata)
}

// ErrCtxUserIDMissing creates an error when a user cannot be found.
func ErrCtxUserIDMissing(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrCtxUserIDMissingMsg, "ctx.userID_not_found", err, logger, metadata)
}

// ErrUserExists creates an error for an already existing user (email or username).
func ErrUserExists(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeConflict, ErrUserExistsMsg, "user_registration_conflict", err, logger, metadata)
}

// ErrAccountLocked creates an error indicating the user's account is locked.
func ErrAccountLocked(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeForbidden, ErrAccountLockedMsg, "account_status_locked", err, logger, metadata)
}

// ErrAccountSuspended creates an error indicating the user's account is suspended.
func ErrAccountSuspended(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeForbidden, ErrAccountSuspendedMsg, "account_status_suspended", err, logger, metadata)
}

// ErrRegistrationFailed creates a generic error for user registration failure.
func ErrRegistrationFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrRegistrationFailedMsg, "user_registration_process_failed", err, logger, metadata)
}

// ErrInvalidPassword creates an error for a password that fails strength requirements or is incorrect.
func ErrInvalidPassword(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidPasswordMsg, "password_validation", err, logger, metadata)
}

// ErrInvalidEmailFormat creates an error for an invalid email address format.
func ErrInvalidEmailFormat(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidEmailFormatMsg, "email_format_validation", err, logger, metadata)
}

// ErrInvalidUsername creates an error for an invalid username format or content.
func ErrInvalidUsername(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidUsernameMsg, "username_validation", err, logger, metadata)
}

// ErrInvalidPhoneNumber creates an error for an invalid phone number format.
func ErrInvalidPhoneNumber(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidPhoneNumberMsg, "phone_number_validation", err, logger, metadata)
}

// ErrInvalidDateFormat creates an error for an invalid date format.
func ErrInvalidDateFormat(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidDateFormatMsg, "date_format_validation", err, logger, metadata)
}

// ErrMissingRequiredField creates an error for missing required fields in a request.
func ErrMissingRequiredField(fieldName string, err error, logger logging.Logger, metadata common.Envelop) *AppError {
	// Using the template constant for the message
	return NewAppError(CodeBadRequest, ErrMissingRequiredFieldMsgTpl(fieldName), "required_fields_check", err, logger, metadata)
}

// ErrMissingUserContext creates an error when user context is missing (e.g., in a request handler).
func ErrMissingUserContext(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnauthorized, ErrMissingUserContextMsg, "user_context_not_found", err, logger, metadata)
}

// ---
// ## Database & Persistence Errors
// These functions create errors related to data storage and retrieval operations.
// ---

// ErrDatabaseConnectionFailed creates an error indicating a failure to connect to the database.
func ErrDatabaseConnectionFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrDatabaseConnectionFailedMsg, "database_connection_failure", err, logger, metadata)
}

// ErrQueryFailed creates an error for a general database query failure.
func ErrQueryFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrQueryFailedMsg, "database_query_failed", err, logger, metadata)
}

// ErrDatabaseTimeout creates an error when a database operation times out.
func ErrDatabaseTimeout(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrDatabaseTimeoutMsg, "database_operation_timeout", err, logger, metadata)
}

// ErrDuplicateEntry creates an error for a duplicate entry in the database.
func ErrDuplicateEntry(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeConflict, ErrDuplicateEntryMsg, "database_duplicate_entry", err, logger, metadata)
}

// ErrForeignKeyViolation creates an error for a foreign key constraint violation.
func ErrForeignKeyViolation(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnprocessable, ErrForeignKeyViolationMsg, "database_foreign_key_violation", err, logger, metadata)
}

// ErrTransactionFailed creates an error for a failed database transaction.
func ErrTransactionFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrTransactionFailedMsg, "database_transaction_failed", err, logger, metadata)
}

// ErrRecordNotFound creates an error when a specific record is not found in the database.
func ErrRecordNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrNotFoundMsg, "database_record_not_found", err, logger, metadata)
}

// ---
// ## Resource Management Errors
// These functions create errors related to lifecycle of resources (creation, update, deletion).
// ---

// ErrResourceNotFound creates an error when a requested resource could not be found.
func ErrResourceNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrNotFoundMsg, "resource_lookup", err, logger, metadata)
}

// ErrResourceCreationFailed creates an error when resource creation fails.
func ErrResourceCreationFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrResourceCreationFailedMsg, "resource_creation_failed", err, logger, metadata)
}

// ErrResourceUpdateFailed creates an error when resource update fails.
func ErrResourceUpdateFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrResourceUpdateFailedMsg, "resource_update_failed", err, logger, metadata)
}

// ErrResourceDeletionFailed creates an error when resource deletion fails.
func ErrResourceDeletionFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrResourceDeletionFailedMsg, "resource_deletion_failed", err, logger, metadata)
}

// ---
// ## File & Storage Errors
// These functions create errors related to file operations and storage.
// ---

// ErrFileNotFound creates an error for a missing file.
func ErrFileNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrNotFoundMsg, "file_lookup", err, logger, metadata)
}

// ErrFileUploadFailed creates an error for a failed file upload.
func ErrFileUploadFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrFileUploadFailedMsg, "file_upload_process_failed", err, logger, metadata)
}

// ErrUnsupportedFileType creates an error for an unsupported file type during upload.
func ErrUnsupportedFileType(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrUnsupportedFileTypeMsg, "file_type_validation", err, logger, metadata)
}

// ErrFileTooLarge creates an error for a file size exceeding the limit.
func ErrFileTooLarge(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrFileTooLargeMsg, "file_size_limit_exceeded", err, logger, metadata)
}

// ErrDiskSpaceFull creates an error when the disk space is full.
func ErrDiskSpaceFull(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrDiskSpaceFullMsg, "disk_space_full", err, logger, metadata)
}

// ErrFileCreateFailed creates an error when the disk space is full.
func ErrFileCreateFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrFileCreateFailedMsg, "os_create_file", err, logger, metadata)
}

// ErrFileCreateFailed creates an error when the disk space is full.
func ErrFileCopyFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrFileCopyFailedMsg, "os_create_file", err, logger, metadata)
}

// ErrFileSystemReadOnly creates an error when the file system is read-only.
func ErrFileSystemReadOnly(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrFileSystemReadOnlyMsg, "file_system_read_only", err, logger, metadata)
}

// ErrDirectoryNotFound creates an error when a directory is not found.
func ErrDirectoryNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrDirectoryNotFoundMsg, "directory_lookup", err, logger, metadata)
}

// ---
// ## Network & External Service Errors
// These functions create errors for issues interacting with external APIs or network connectivity.
// ---

// ErrNetworkUnreachable creates a network error when the network is unreachable.
func ErrNetworkUnreachable(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrNetworkUnreachableMsg, "network_unreachable", err, logger, metadata)
}

// ErrDNSResolutionFailed creates a network error for a DNS resolution failure.
func ErrDNSResolutionFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrDNSResolutionFailedMsg, "dns_resolution_failed", err, logger, metadata)
}

// ErrExternalServiceUnavailable creates an error when an external service is unavailable.
func ErrExternalServiceUnavailable(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrExternalAPIServiceUnavailableMsg, "external_service_unavailable", err, logger, metadata)
}

// ErrExternalAPIRequestFailed creates an error for a failed external API request.
func ErrExternalAPIRequestFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrExternalAPIRequestFailedMsg, "external_api_request_failed", err, logger, metadata)
}

// ErrRateLimitExceeded creates an error when a rate limit is exceeded.
func ErrRateLimitExceeded(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeRateLimitExceeded, ErrTooManyRequestsMsg, "rate_limit_exceeded", err, logger, metadata)
}

// ErrPaymentGatewayTimeout creates an error when a payment gateway request times out.
func ErrPaymentGatewayTimeout(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrPaymentGatewayTimeoutMsg, "payment_gateway_timeout", err, logger, metadata)
}

// ErrInvalidPaymentMethod creates an error for an invalid payment method.
func ErrInvalidPaymentMethod(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidPaymentMethodMsg, "payment_method_validation", err, logger, metadata)
}

// ErrTransactionDeclined creates an error when a transaction is declined.
func ErrTransactionDeclined(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodePaymentRequired, ErrTransactionDeclinedMsg, "transaction_declined", err, logger, metadata)
}

// ErrInsufficientFunds creates an error for insufficient funds.
func ErrInsufficientFunds(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodePaymentRequired, ErrInsufficientFundsMsg, "insufficient_funds", err, logger, metadata)
}

// ErrBillingInfoMissing creates an error when billing information is missing.
func ErrBillingInfoMissing(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrBillingInfoMissingMsg, "billing_info_missing", err, logger, metadata)
}

// ---
// ## Request & Validation Errors
// These functions create errors related to the HTTP request itself or its data.
// ---

// ErrBadRequest creates an error for a malformed or invalid HTTP request.
func ErrBadRequest(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrBadRequestMsg, "http_request_malformed", err, logger, metadata)
}

// ErrInvalidRequestPayload creates an error for a malformed or invalid request body payload.
func ErrInvalidRequestPayload(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidRequestPayloadMsg, "request_payload_parsing_failed", err, logger, metadata)
}

// ErrInvalidIDParameter creates an error for an invalid ID parameter in the request.
func ErrInvalidIDParameter(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidIDParameterMsg, "id_parameter_validation", err, logger, metadata)
}

// ErrInvalidContentType creates an error for an unsupported content type.
func ErrInvalidContentType(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeUnsupportedMediaType, ErrInvalidContentTypeMsg, "content_type_validation", err, logger, metadata)
}

// ErrFieldRequired creates an error for a missing required field.
func ErrFieldRequired(fieldName string, err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrMissingRequiredFieldMsgTpl(fieldName), "field_required_validation", err, logger, metadata)
}

// ErrRequestBodyTooLarge creates an error for an oversized request body.
func ErrRequestBodyTooLarge(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeRequestBodyTooLarge, ErrRequestBodyTooLargeMsg, "request_body_size_exceeded", err, logger, metadata)
}

// ErrRequestTimeout creates an error when a client request times out.
func ErrRequestTimeout(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeRequestTimeout, ErrOperationTimedOutMsg, "client_request_timeout", err, logger, metadata)
}

// ErrRequestValidation creates an error when client-side request validation fails.
func ErrRequestValidation(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	// Fix: Changed Code to CodeBadRequest/CodeUnprocessable and context string
	return NewAppError(CodeUnprocessable, ErrValidationMsg, "request_validation_failed", err, logger, metadata)
}

// ---
// ## JSON Specific Errors
// These functions create errors related to JSON encoding and decoding.
// ---

// ErrJSONDecodeFailed creates an error when JSON decoding fails.
func ErrJSONDecodeFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrJSONDecodeFailedMsg, "json_decode_failed", err, logger, metadata)
}

// ErrJSONEncodeFailed creates an error when JSON encoding fails (typically for response).
func ErrJSONEncodeFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrJSONEncodeFailedMsg, "json_encode_failed", err, logger, metadata)
}

// ErrInvalidJSONStructure creates an error for an invalid JSON structure.
func ErrInvalidJSONStructure(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrInvalidJSONStructureMsg, "json_structure_invalid", err, logger, metadata)
}

// ---
// ## Movie, Actor, Genre Errors
// These functions create errors specific to your movie application domain.
// ---

// ErrInvalidMovieID creates an error for an invalid movie ID.
func ErrInvalidMovieID(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrInvalidMovieIDMsg, "movie_id_validation", err, logger, metadata)
}

// ErrMovieNotFound creates an error when a movie is not found.
func ErrMovieNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrMovieNotFoundMsg, "movie_lookup", err, logger, metadata)
}

// ErrMovieCreateFailed creates an error when movie creation fails.
func ErrMovieCreateFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrMovieCreateFailedMsg, "movie_creation_failed", err, logger, metadata)
}

// ErrMovieUpdateFailed creates an error when movie update fails.
func ErrMovieUpdateFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrMovieUpdateFailedMsg, "movie_update_failed", err, logger, metadata)
}

// ErrMovieDeleteFailed creates an error when movie deletion fails.
func ErrMovieDeleteFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrMovieDeleteFailedMsg, "movie_deletion_failed", err, logger, metadata)
}

// ErrMovieAlreadyExists creates an error for an existing movie.
func ErrMovieAlreadyExists(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeConflict, ErrMovieAlreadyExistsMsg, "movie_already_exists", err, logger, metadata)
}

// ErrMovieRatingOutOfBounds creates an error for an invalid movie rating.
func ErrMovieRatingOutOfBounds(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrMovieRatingOutOfBoundsMsg, "movie_rating_validation", err, logger, metadata)
}

// ErrMovieCastNotFound creates an error for missing movie cast information.
func ErrMovieCastNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrMovieCastNotFoundMsg, "movie_cast_lookup", err, logger, metadata)
}

// ErrMovieGenreNotFound creates an error for missing movie genre.
func ErrMovieGenreNotFound(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrMovieGenreNotFoundMsg, "movie_genre_lookup", err, logger, metadata)
}

// ErrInvalidGenreID creates an error for an invalid genre ID.
func ErrInvalidGenreID(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrInvalidGenreIDMsg, "genre_id_validation", err, logger, metadata)
}

// ErrMovieReviewFailed creates an error when posting a movie review fails.
func ErrMovieReviewFailed(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrMovieReviewFailedMsg, "movie_review_submission_failed", err, logger, metadata)
}

// ErrInvalidActorID creates an error for an invalid actor ID.
func ErrInvalidActorID(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotFound, ErrInvalidActorIDMsg, "actor_id_validation", err, logger, metadata)
}

// ---
// ## Miscellaneous & System Errors
// These functions handle general application issues or infrastructure problems.
// ---

// ErrInternalServer creates a generic internal server error.
func ErrInternalServer(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrInternalServerMsg, "internal_server_error_general", err, logger, metadata)
}

// NoLoggerErr creates an error when the logger instance is unavailable.
func NoLoggerErr() *AppError {
	return NewAppError(CodeInternal, ErrNoLoggerMsg, "app_logger_unavailable", nil, nil, nil)
}

// ErrRedisNotConfigured creates an error when the Redis URL is not set in environment.
func ErrRedisNotConfigured(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrRedisNotConfiguredMsg, "redis_config_missing", err, logger, metadata)
}

// ErrFailedJSONResWrite creates an error when writing JSON response fails.
func ErrFailedJSONResWrite(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrFailedJSONResWriteMsg, "json_response_write_failed", err, logger, metadata)
}

// ErrMissingJWTSecret creates an internal error for a missing JWT secret configuration.
func ErrMissingJWTSecret(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrMissingJWTSecretConfigMsg, "jwt_secret_config_missing", err, logger, metadata)
}

// ErrTokenHashFailedInternal creates an internal error for a token hash operation failure.
func ErrTokenHashFailedInternal(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrTokenHashFailedInternalMsg, "token_hash_operation_failed", err, logger, metadata)
}

// ErrHashFunctionUnavailable creates an internal error for an unavailable hash function.
func ErrHashFunctionUnavailable(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrHashFunctionUnavailableMsg, "crypto_hash_function_unavailable", err, logger, metadata)
}

// ErrUnexpectedSystemError creates a general unexpected system error.
func ErrUnexpectedSystemError(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrUnexpectedSystemErrorMsg, "unexpected_system_error", err, logger, metadata)
}

// ErrMaintenanceModeActive creates an error when the system is in maintenance mode.
func ErrMaintenanceModeActive(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeServiceUnavailable, ErrMaintenanceModeActiveMsg, "system_maintenance_mode", err, logger, metadata)
}

// ErrFeatureNotImplemented creates an error for a feature that is not yet implemented.
func ErrFeatureNotImplemented(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotImplemented, ErrFeatureNotImplementedMsg, "feature_not_implemented", err, logger, metadata)
}

// ErrUnsupportedOperation creates an error for an unsupported operation.
func ErrUnsupportedOperation(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeNotImplemented, ErrUnsupportedOperationMsg, "unsupported_operation", err, logger, metadata)
}

// ErrUnsupportedLanguage creates an error for an unsupported language or locale.
func ErrUnsupportedLanguage(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeBadRequest, ErrUnsupportedLanguageMsg, "language_locale_unsupported", err, logger, metadata)
}

// QueryNotFoundError creates an error when a SQL query definition is missing internally.
func QueryNotFoundError(queryKey string, logger logging.Logger, metadata common.Envelop) *AppError {
	op := fmt.Sprintf("store.QueryNotFound:%s", queryKey)
	metadata["query_key"] = queryKey
	return NewAppError(CodeInternal, ErrInternalServerMsg, op, nil, logger, metadata)
}

// ErrDatabaseOpFailed is a general factory for database operation failures.
func ErrDatabaseOpFailed(code string, msg string, op string, originalErr error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(code, msg, op, originalErr, logger, metadata)
}

// ErrUniqueViolation is a general factory for database operation failures.
func ErrUniqueViolation(err error, op string, log logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeConflict, ErrConflictMsg, op, err, log, metadata)
}

// ErrRedisURLNotSet is a general factory for database operation failures.
func ErrRedisURLNotSet(err error, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(CodeInternal, ErrRedisNotConfiguredMsg, "config", err, logger, metadata)
}

// ErrMovieNotFound is a specific factory for movie not found errors.
// func ErrMovieNotFound(originalErr error, logger logging.Logger, metadata common.Envelop) *apperror.AppError {
// 	return apperror.NewAppError(
// 		apperror.CodeMovieNotFound, // Use CodeMovieNotFound string constant
// 		apperror.ErrMovieNotFoundMsg,
// 		"repository.movie_not_found",
// 		originalErr,
// 		logger,
// 		metadata,
// 	)
// }

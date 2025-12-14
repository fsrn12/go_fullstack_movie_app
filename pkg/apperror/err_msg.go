package apperror

import "fmt"

// =====================================
// GENERAL & COMMON CLIENT MESSAGES
// These messages cover broad categories that apply across the application.
// They are concise and guide the user where possible.
// =====================================
const (
	ErrBadRequestMsg           = "Your request was malformed or missing required information. Please check it and try again."
	ErrConflictMsg             = "The operation couldn't be completed due to a conflict with the current state of the resource. This might happen if you're trying to create something that already exists."
	ErrInternalServerMsg       = "An unexpected error occurred. We're working to fix it. Please try again later."
	ErrInvalidJSONStructureMsg = "Invalid JSON structure. Please check the format."
	ErrMethodNotAllowedMsg     = "The requested method is not allowed for this resource."
	ErrNotFoundMsg             = "The requested resource could not be found."                 // Primary "not found" message, covers general resource, records, and files.
	ErrOperationTimedOutMsg    = "The operation took too long to complete. Please try again." // More general than request/query timeouts, applies broadly.
	ErrServiceUnavailableMsg   = "The service is temporarily unavailable. We're working to restore it. Please try again later."
	ErrTooManyRequestsMsg      = "You've sent too many requests recently. Please wait a bit and try again."
	ErrValidationMsg           = "There was an issue with the data you provided. Please check the input fields and try again." // General validation error. Specific field errors are best conveyed via AppError.Metadata.
)

// =====================================
// AUTHENTICATION & AUTHORIZATION CLIENT MESSAGES
// Clear, actionable messages for identity and access issues.
// =====================================
const (
	ErrCSRFTokenMissingMsg      = "A security token (CSRF) is missing. Please refresh the page and try again."
	ErrForbiddenMsg             = "You don't have permission to access this resource or perform this action."
	ErrInvalidAPIKeyMsg         = "The API key provided is invalid or has expired."
	ErrInvalidAuthMsg           = "Your authentication details are invalid. Please check your credentials and try again." // Covers general auth, invalid header, invalid JWT token.
	ErrInvalidTokenSignatureMsg = "The token's signature is invalid. This token might have been tampered with."
	ErrMissingAuthMsg           = "An authentication token is required but was not provided."
	ErrTokenExpiredMsg          = "Your session has expired. Please log in again." // Covers JWT, OAuth, and general session expiry.
	ErrTokenMalformedMsg        = "The provided token is malformed. Please ensure it's in the correct format."
	ErrTokenNotActiveMsg        = "Your token is not yet active. Please try again later." // Specific for 'nbf' claim.
	ErrTokenNotFoundMsg         = "Authentication token not found or is invalid."         // Clarified message for distinction from ErrMissingAuth
	ErrUnauthorizedMsg          = "Authentication is required to access this resource. Please log in."
)

// =====================================
// USER ACCOUNT & PROFILE CLIENT MESSAGES
// Specific messages for user lifecycle and account states.
// =====================================
const (
	ErrAccountLockedMsg      = "Your account has been temporarily locked due to too many failed login attempts. Please try again later or contact support."
	ErrAccountSuspendedMsg   = "Your account has been suspended. Please contact support for more information."
	ErrInvalidDateFormatMsg  = "The date format is invalid. Please use a recognized date format."
	ErrInvalidEmailFormatMsg = "The email address you entered is not valid. Please check the format and try again."
	ErrInvalidNameMsg        = "Name is required and must be between 2 and 32 characters. It can only contain letters, hyphen, underscore, apostrophe, and period. Examples:'James T. Kirk', 'Jean-Luc Picard' and 'Rick O'Connell'."
	ErrInvalidPasswordMsg    = "Your password must be at least 8 characters long and include a mix of uppercase letters, lowercase letters, numbers, and special characters."
	ErrInvalidPhoneNumberMsg = "The phone number format is invalid. Please check and try again."
	ErrInvalidUsernameMsg    = "The username you entered is invalid. Usernames can only contain letters, numbers, and underscores, and must be between 3 and 20 characters."
	ErrMissingUserContextMsg = "User context missing from request."
	ErrRegistrationFailedMsg = "We couldn't complete your registration. Please try again, or contact support if the problem persists."
	ErrUserExistsMsg         = "A user with this email address or username already exists. Please try logging in or use a different email/username."
	ErrUserNotFoundMsg       = "We couldn't find a user matching the provided details."
	ErrCtxUserMissingMsg     = "User not found in context"
	ErrCtxUserIDMissingMsg   = "UserID not found in context"
)

func ErrMissingRequiredFieldMsg(field string) string {
	return fmt.Sprintf("The '%s' field is required. Please provide a valid value.", field)
}

// =====================================
// RESOURCE & DATABASE CLIENT MESSAGES
// Messages for issues during data storage and retrieval.
// =====================================
const (
	ErrDatabaseConnectionFailedMsg = "We're unable to connect to our database right now. Please try again in a moment."
	ErrDatabaseTimeoutMsg          = "The database operation timed out. Please try again."
	ErrDuplicateEntryMsg           = "The record you're trying to create already exists. Please check for duplicates."
	ErrForeignKeyViolationMsg      = "This action cannot be completed because it's linked to another record that doesn't exist or is currently in use."
	ErrQueryFailedMsg              = "Failed to process the database query." // More general for logging, client might get 500 or more specific based on context
	ErrResourceCreationFailedMsg   = "We couldn't create the resource. Please check your input and try again."
	ErrResourceDeletionFailedMsg   = "We couldn't delete the resource. It might not exist or be linked to other data."
	ErrResourceUpdateFailedMsg     = "We couldn't update the resource. It might have been deleted or changed by someone else."
	ErrTransactionFailedMsg        = "The operation couldn't be completed due to a transactional error. Please try again."
)

// =====================================
// REQUEST INPUT & FORMATTING CLIENT MESSAGES
// Messages for problems with the structure or format of the request.
// =====================================
const (
	ErrFieldRequiredMsg         = "A required field is missing. Please ensure all necessary fields are provided." // General "field required"
	ErrInvalidContentTypeMsg    = "The 'Content-Type' header is missing or unsupported. Please use 'application/json'."
	ErrInvalidIDParameterMsg    = "The ID you provided is invalid. Please ensure it's in the correct format."
	ErrInvalidRequestPayloadMsg = "The request body is malformed or invalid."                           // Covers general malformed body, invalid JSON structure, bad data.
	ErrJSONDecodeFailedMsg      = "Failed to process the request's JSON data. Please check its format." // Important for logging, but also somewhat client-friendly.
	ErrJSONEncodeFailedMsg      = "Failed to prepare the response data. This is an internal issue."     // More for logging, but a client *might* see it if response writing fails.
	ErrRequestBodyTooLargeMsg   = "The request body is too large. Please reduce its size and try again."
)

// =====================================
// FILE & STORAGE CLIENT MESSAGES
// Messages related to file upload, storage, and access.
// =====================================
const (
	ErrDirectoryNotFoundMsg   = "The specified directory doesn't exist."
	ErrDiskSpaceFullMsg       = "We're experiencing storage issues. Please try again later."
	ErrFileTooLargeMsg        = "The file size exceeds our allowed limit. Please upload a smaller file."
	ErrFileNotFoundMsg        = "The file you're looking for couldn't be found."
	ErrFileUploadFailedMsg    = "We couldn't upload your file. Please try again."
	ErrFileSystemReadOnlyMsg  = "The file system is currently in read-only mode. Please try again later."
	ErrUnsupportedFileTypeMsg = "The file type you're trying to upload is not supported."
	ErrFileCreateFailedMsg    = "Failed to create file"
	ErrFileCopyFailedMsg      = "Failed to copy file content"
)

// =====================================
// NETWORK & EXTERNAL SERVICE CLIENT MESSAGES
// Messages for issues interacting with external APIs or network connectivity.
// =====================================
const (
	ErrBillingInfoMissingMsg            = "Billing information is missing. Please provide your billing details."
	ErrDNSResolutionFailedMsg           = "We couldn't resolve the necessary network address. Please try again."
	ErrExternalAPIRequestFailedMsg      = "We encountered an issue communicating with an external service. Please try again later."
	ErrExternalAPIServiceUnavailableMsg = "An external service is currently unavailable. Please try again in a moment."
	ErrInsufficientFundsMsg             = "Your account has insufficient funds to complete this transaction."
	ErrInvalidPaymentMethodMsg          = "The payment method you provided is invalid. Please check your details."
	ErrNetworkUnreachableMsg            = "We're unable to connect to the network. Please check your internet connection and try again." // General network issue.
	ErrPaymentGatewayTimeoutMsg         = "Your payment couldn't be processed in time. Please try again."
	ErrTransactionDeclinedMsg           = "Your transaction was declined. Please check your payment details or contact your bank."
)

// =====================================
// CRYPTOGRAPHY & SECURITY CLIENT MESSAGES
// Messages for encryption, hashing, and token integrity.
// =====================================
const (
	ErrDecryptionFailedMsg        = "We encountered an issue decrypting data. Please try again later."
	ErrEncryptionFailedMsg        = "We encountered an issue encrypting data. Please try again later."
	ErrHashFunctionUnavailableMsg = "A required security function is currently unavailable."
	ErrInvalidKeyMsg              = "The cryptographic key used is invalid or corrupted."
	ErrInvalidKeyTypeMsg          = "The cryptographic key type is not supported."
	ErrInvalidTokenMsg            = "The provided token is invalid. Please ensure it's correct." // Covers various invalid token states.
)

// =====================================
// CACHE & PERFORMANCE CLIENT MESSAGES
// Messages for caching issues or general performance problems.
// =====================================
const (
	ErrCacheMissMsg          = "The requested data was not found in cache. Retrieving from primary source." // Can be informative for performance insights.
	ErrCacheStorageFailedMsg = "Failed to store data in cache."
	ErrCacheTimeoutMsg       = "The cache operation timed out. Data might be stale."
)

// =====================================
// SYSTEM & MAINTENANCE CLIENT MESSAGES
// Messages for internal system state or features visible to the user.
// =====================================
const (
	ErrFeatureNotImplementedMsg = "This feature is not yet available. Please check back later."
	ErrMaintenanceModeActiveMsg = "Our system is currently undergoing maintenance. Please try again shortly."
	ErrUnexpectedSystemErrorMsg = "An unexpected system error occurred. We're investigating the issue."
	ErrUnsupportedLanguageMsg   = "The requested language or locale is not supported."
	ErrUnsupportedOperationMsg  = "This operation is not supported for the requested resource."
)

// =====================================
// INFRASTRUCTURE & CONFIGURATION INTERNAL MESSAGES
// These messages are primarily for logging and debugging by developers/ops.
// They should typically NOT be sent directly to clients as `AppError.Message`.
// Instead, use them in AppError.Op or AppError.Metadata.
// =====================================
const (
	ErrFailedJSONResWriteMsg      = "Failed to write JSON response to client. Possible broken connection or internal marshalling issue."
	ErrMissingJWTSecretConfigMsg  = "Internal server error: JWT secret configuration is missing." // Clarified from previous "The provided JWT token is invalid"
	ErrNoLoggerMsg                = "Logger instance is unavailable for error reporting."
	ErrRedisNotConfiguredMsg      = "Redis connection URL is not set in environment configuration."
	ErrTokenHashFailedInternalMsg = "Internal: Failed to compute or verify token hash." // Renamed for clarity on internal use.
)

// =====================================
// DOMAIN-SPECIFIC CLIENT MESSAGES (e.g., Movie, Actor, Genre)
// These messages are highly specific to your application's business logic.
// =====================================
const (
	ErrInvalidActorIDMsg         = "We couldn't find the actor with that ID. Please try again with a valid actor ID."
	ErrInvalidGenreIDMsg         = "We couldn't find the genre with that ID. Please try again with a valid genre ID."
	ErrInvalidMovieIDMsg         = "We couldn't find the movie with that ID. Please try again with a valid movie ID."
	ErrMovieAlreadyExistsMsg     = "A movie with similar details already exists."
	ErrMovieCastNotFoundMsg      = "We couldn't find the cast for this movie."
	ErrMovieCreateFailedMsg      = "We couldn't create the movie. Please check your input and try again."
	ErrMovieDeleteFailedMsg      = "We couldn't delete the movie. It might not exist or be linked to other data."
	ErrMovieGenreNotFoundMsg     = "We couldn't find the genre for this movie."
	ErrMovieNotFoundMsg          = "The movie you are looking for could not be found."
	ErrMovieRatingOutOfBoundsMsg = "Movie ratings must be between 0 and 10."
	ErrMovieReviewFailedMsg      = "We couldn't post your movie review. Please try again."
	ErrMovieUpdateFailedMsg      = "We couldn't update the movie. It might have been deleted or changed by someone else."
)

func ErrMissingRequiredFieldMsgTpl(field string) string {
	return fmt.Sprintf("The '%s' field is required. Please provide a valid value.", field)
}

// // =============================
// // AUTHENTICATION & AUTHORIZATION ERRORS
// // =============================

// const (
// 	ErrUnauthorizedAccessMsg = "Authentication required for access to this resource"
// 	ErrForbiddenAccessMsg    = "Access to the resource is forbidden"
// 	ErrForbiddenMsg          = "You do not have permission to perform this action."
// 	ErrInvalidAuthHeaderMsg  = "Authentication header is malformed or missing"
// 	ErrMissingAuthHeaderMsg  = "Authentication header is missing"
// 	ErrMissingJWTSecretMsg   = "The provided JWT token is invalid"
// 	ErrInvalidJWTTokenMsg    = "The provided JWT token is invalid"
// 	ErrTokenExpiredMsg       = "JWT token has expired"
// 	ErrTokenInvalidMsg       = "JWT token is invalid or malformed"
// 	ErrMissingJWTTokenMsg    = "Required JWT token is missing from request"
// 	ErrMissingTokenMsg       = "Required token is missing from request"
// 	ErrMalformedTokenMsg     = "The JWT token is malformed and cannot be parsed"
// 	ErrOAuthTokenExpiredMsg  = "OAuth token has expired"
// 	ErrTokenHashFailedMsg    = "Failed to retrieve token hash"
// 	ErrMethodNotAllowedMsg   = "Method not allowed"
// )

// // =============================
// // USER REGISTRATION / MANAGEMENT ERRORS
// // =============================

// const (
// 	ErrUserNotFoundMsg                     = "user not found in the system"
// 	ErrUserIDNotFoundMsg                   = "userID not found"
// 	ErrUserAlreadyExistsMsg                = "user already exists"
// 	ErrEmailAlreadyRegisteredMsg           = "email already in use, please log in"
// 	ErrMissingUserContextMsg               = "user context is missing or invalid"
// 	ErrAdminOnlyResourceMsg                = "access restricted to administrators only"
// 	ErrAccountLockedMsg                    = "your account has been locked"
// 	ErrAccountSuspendedMsg                 = "your account has been suspended"
// 	ErrSessionExpiredMsg                   = "your session has expired"
// 	ErrInsufficientPermissionsMsg          = "insufficient permissions to perform this action"
// 	ErrInvalidUsernameMsg                  = "invalid username"
// 	ErrInvalidEmailMsg                     = "invalid email"
// 	ErrInvalidEmailFormatMsg               = "invalid email format"
// 	ErrPasswordStrengthMsg                 = "password does not meet strength requirements"
// 	ErrInvalidPhoneNumberMsg               = "invalid phone number format, please check and try again"
// 	ErrInvalidDateFormatMsg                = "invalid date format"
// 	ErrNameRequiredMsg                     = "name is required and must be between 2 and 50 characters"
// 	ErrUsernameAlreadyTakenMsg             = "username already taken"
// 	ErrUsernameRequiredMsg                 = "username is required and must be between 2 and 50 characters"
// 	ErrUsernameLengthInvalidMsg            = "username must be between 3 and 20 characters"
// 	ErrInvalidNameMsg                      = "name is required, must be 2 to 32 letters only"
// 	ErrUsernameRegexFailMsg                = "username can only contain letters, numbers, and underscores"
// 	ErrEmailRequiredMsg                    = "email is required"
// 	ErrPasswordInvalidMsg                  = "password is invalid, please ensure it meets the strength requirements"
// 	ErrHashingFailedMsg                    = "Failed to hash"
// 	ErrPasswordMissingMsg                  = "Password field value is missing or malformed"
// 	ErrWeakPasswordMsg                     = "password must be at least 8 characters long and include uppercase, lowercase, number, and special character"
// 	ErrRegistrationDataMsg                 = "registration data is malformed, missing, or invalid"
// 	ErrRegistrationFailedMsg               = "user registration failed"
// 	ErrUserRegistrationValidationFailedMsg = "user registration validation failed"
// 	ErrMandatoryFieldsMissingMsg           = `mandatory fields missing:
// * name is required
// * email is required
// * password is required`
// )

// // =============================
// // RESOURCE & DATABASE ERRORS
// // =============================

// const (
// 	ErrResourceNotFoundMsg         = "The requested resource could not be found."
// 	ErrRecordNotFoundMsg           = "no matching record found in the database"
// 	ErrResourceCreationFailedMsg   = "failed to create the resource"
// 	ErrResourceUpdateFailedMsg     = "failed to update the resource"
// 	ErrResourceDeletionFailedMsg   = "failed to delete the resource"
// 	ErrDatabaseConnectionFailedMsg = "failed to connect to the database"
// 	ErrQueryTimeoutMsg             = "query timed out"
// 	ErrDuplicateEntryMsg           = "A resource with similar details already exists."
// 	ErrForeignKeyViolationMsg      = "foreign key constraint violation"
// 	ErrTransactionFailedMsg        = "transaction failed"
// 	ErrQueryFailedMsg              = "failed to query"
// )

// // =============================
// // REQUEST & VALIDATION ERRORS
// // =============================

// const (
// 	ErrBadRequestMsg               = "The request is malformed or missing required fields. Please check and try again."
// 	ErrInvalidRequestPayloadMsg    = "Request payload is malformed or invalid"
// 	ErrInvalidRequestBodyFormatMsg = "Request body format is malformed or invalid"
// 	ErrInvalidRequestDataMsg       = "Request body data is malformed or invalid"
// 	ErrConflictMsg                 = "A conflict occurred while processing the request."
// 	ErrValidationMsg               = "The request data is invalid."
// 	ErrRequestTimeoutMsg           = "Request timed-out. Please try again later."
// 	ErrInvalidRequestMsg           = "The request is invalid"
// 	ErrInvalidIDParameterMsg       = "The provided ID parameter is invalid"
// 	ErrInvalidContentTypeMsg       = "Invalid content type - expected application/json"
// 	ErrFieldRequiredMsg            = "Field is required"
// 	ErrInvalidAPIKeyMsg            = "Invalid API key"
// 	ErrRequestBodyTooLargeMsg      = "Request body too large. Please try again"
// 	ErrMissingRequiredParamMsg     = "Missing required params in the request"
// 	ErrRequiredFieldMissingMsg     = "Missing required field in the request"
// 	ErrJSONDecodeFailedMsg         = "Failed to decode JSON data"
// 	ErrJSONEncodeFailedMsg         = "Failed to encode JSON data"
// 	ErrInvalidJSONStructureMsg     = "Invalid JSON structure. Please check the format"
// 	ErrJSONFieldRequiredMsg        = "Missing required field in JSON data. Please check and provide the necessary field"
// )

// // =============================
// // FILE & STORAGE ERRORS
// // =============================

// const (
// 	ErrFileNotFoundMsg        = "The requested file was not found"
// 	ErrFileUploadFailedMsg    = "File upload failed"
// 	ErrUnsupportedFileTypeMsg = "Unsupported file type"
// 	ErrFileTooLargeMsg        = "File exceeds the maximum allowed size"
// 	ErrDiskSpaceFullMsg       = "Disk space is full"
// 	ErrFileSystemReadOnlyMsg  = "File system is read-only"
// 	ErrDirectoryNotFoundMsg   = "Directory not found"
// )

// // =============================
// // NETWORK & EXTERNAL SERVICE ERRORS
// // =============================

// const (
// 	ErrTimeoutWaitingForResponseMsg     = "Timed out waiting for response"
// 	ErrDNSResolutionFailedMsg           = "DNS resolution failed"
// 	ErrServiceUnavailableMsg            = "The service is temporarily unavailable. Please try again later."
// 	ErrNetworkUnreachableMsg            = "Network is unreachable"
// 	ErrExternalAPIServiceUnavailableMsg = "External API service is unavailable"
// 	ErrExternalAPIRequestFailedMsg      = "External API request failed"
// 	ErrRateLimitExceededMsg             = "Rate limit exceeded"
// 	ErrPaymentGatewayTimeoutMsg         = "Payment gateway request timed out"
// 	ErrInvalidPaymentMethodMsg          = "Invalid payment method"
// 	ErrTransactionDeclinedMsg           = "Transaction was declined"
// 	ErrInsufficientFundsMsg             = "Insufficient funds"
// 	ErrBillingInfoMissingMsg            = "Billing information is missing"
// 	ErrInvalidPasswordStrengthMsg       = "Password must be at least 8 characters, and include a mix of numbers, letters, and special characters."
// )

// // =============================
// // SECURITY & CRYPTO ERRORS
// // =============================

// const (
// 	ErrInvalidKeyMsg                = "Provided key is invalid"
// 	ErrInvalidKeyTypeMsg            = "Provided key type is invalid"
// 	ErrHashFunctionUnavailableMsg   = "Requested hash function is not available"
// 	ErrFailedToRetrieveTokenHashMsg = "Failed to retrieve refresh token hash from storage"
// 	ErrEncryptionFailedMsg          = "Encryption failed"
// 	ErrDecryptionFailedMsg          = "Decryption failed"
// 	ErrInvalidTokenSignatureMsg     = "Invalid token signature"
// 	ErrCSRFTokenMissingMsg          = "CSRF token is missing"
// 	ErrInvalidTokenMsg              = "Invalid token"
// 	ErrGenerateTokenFailedMsg       = "Failed to generate tokens"
// 	ErrRefreshTokenFailedMsg        = "Failed to create refresh token"
// 	ErrRefreshTokenNotFoundMsg      = "Refresh token not found or expired"
// 	ErrRefreshTokenNotSavedMsg      = "Refresh token not saved"
// 	ErrRefreshTokenInvalidMsg       = "Refresh token malformed or invalid"
// 	ErrRefreshTokenHashMisMatchMsg  = "Refresh token hash mismatch"
// 	ErrTokenNotActiveMsg            = "Token is not active"
// 	ErrFailedToGenerateTokenMsg     = "Failed to generate token"
// 	ErrTokenNotFoundMsg             = "Token not found"
// 	ErrInvalidJWTFormatMsg          = "Invalid JWT format"
// 	ErrInvalidJWTClaimsMsg          = "Invalid JWT claims"
// 	ErrTokenClaimsInvalidMsg        = "Token claims malformed or invalid"
// 	ErrTokenClaimsMissingMsg        = "Token claims missing"
// 	ErrRefreshTokenDeleteFailedMsg  = "Failed to delete refresh token"
// )

// // =============================
// // CACHE & PERFORMANCE ERRORS
// // =============================

// const (
// 	ErrCacheMissMsg            = "Cache miss"
// 	ErrCacheTimeoutMsg         = "Cache timeout"
// 	ErrCacheStorageFailedMsg   = "Cache storage failed"
// 	ErrOperationTimedOutMsg    = "Operation timed out"
// 	ErrTooManyRequestsMsg      = "Too many requests, please try again later"
// 	ErrRetryOperationFailedMsg = "Retry operation failed"
// )

// // =============================
// // SYSTEM & MAINTENANCE ERRORS
// // =============================

// const (
// 	ErrUnrecognizedErrorMsg     = "An unrecognized error occurred"
// 	ErrUnexpectedSystemErrorMsg = "Unexpected system error"
// 	ErrMaintenanceModeActiveMsg = "The system is in maintenance mode"
// 	ErrFeatureNotImplementedMsg = "This feature is not yet implemented"
// 	ErrUnsupportedOperationMsg  = "This operation is unsupported"
// 	ErrUnsupportedLanguageMsg   = "Unsupported language/locale"
// )

// // =============================
// // GENERAL & CONFIGURATION ERRORS
// // =============================

// const (
// 	NoLoggerErrMsg        = "Failed to get logger"
// 	ErrRedisURLNotSetMsg  = "redis_url not set in env"
// 	ErrInternalServerMsg  = "An unexpected error occurred. Please try again later."
// 	ErrFailedJSONResWrite = "Failed to write to JSON response"
// )

// // =============================
// // MOVIE, ACTOR, GENRE ERRORS
// // =============================

// const (
// 	ErrInvalidMovieIDMsg          = "Oops! We couldn't find the movie with the provided ID. Please try again with a valid ID, like '13'."
// 	ErrInvalidActorIDMsg          = "Oops! We couldn't find the actor with the provided ID. Please try again with a valid ID, like '13'."
// 	ErrInvalidGenreIDMsg          = "Oops! We couldn't find the genre with the provided ID. Please try again with a valid ID, like '13'."
// 	ErrMovieNotFoundMsg           = "Sorry! We couldn't find the movie you were looking for."
// 	ErrTopMoviesResponseFailedMsg = "failed to write JSON response for top movies"
// 	ErrTopMovieFailedMsg          = "Failed to get top movies"
// 	ErrMovieCreateFailedMsg       = "Failed to create movie"
// 	ErrMovieUpdateFailedMsg       = "Failed to update movie"
// 	ErrMovieDeleteFailedMsg       = "Failed to delete movie"
// 	ErrMovieAlreadyExistsMsg      = "Movie already exists"
// 	ErrMovieRatingOutOfBoundsMsg  = "Movie can only be rated from 0 to 10"
// 	ErrMovieCastNotFoundMsg       = "Failed to find movie cast"
// 	ErrMovieGenreNotFoundMsg      = "Failed to find genre for the movie"
// 	ErrMovieReviewFailedMsg       = "Failed to post the review for the movie"
// )

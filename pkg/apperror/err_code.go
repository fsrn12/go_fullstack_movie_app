package apperror

import "net/http"

// =====================================
// HTTPStatus maps application error codes (string) to standard HTTP status codes (int).
// This function acts as the bridge between your internal error identification
// and the external HTTP response.
// =====================================
func HTTPStatus(code any) int {
	switch c := code.(type) {
	case string:
		switch c {
		// 2xx Success Codes
		case CodeOK:
			return http.StatusOK
		case CodeAccepted:
			return http.StatusAccepted
		case CodeCreated:
			return http.StatusCreated
		case CodeNoContent:
			return http.StatusNoContent

		// 4xx Client Error Codes
		case CodeBadRequest:
			return http.StatusBadRequest
		case CodeUnauthorized, CodeAccountLocked, CodeAccountSuspended, CodeInvalidAuth,
			CodeInvalidAPIKey, CodeInvalidJWT, CodeInvalidToken, CodeInvalidTokenSignature,
			CodeMissingToken, CodeOAuthError, CodeOAuthTokenExpired, CodeSessionExpired,
			CodeTokenExpired, CodeTokenMalformed, CodeTokenNotActive: // Grouped Authentication Errors
			return http.StatusUnauthorized
		case CodeForbidden, CodeAdminOnlyResource, CodeCSRFTokenMissing, CodeInsufficientPermissions: // Grouped Authorization Errors
			return http.StatusForbidden
		case CodeConflict, CodeDuplicateEntry, CodeMovieAlreadyExists: // Grouped Conflict Errors
			return http.StatusConflict
		case CodeNotFound, CodeUserNotFound, CodeFileNotFound, CodeDirectoryNotFound,
			CodeInvalidMovieID, CodeInvalidActorID, CodeInvalidGenreID, CodeMovieNotFound,
			CodeMovieCastNotFound, CodeMovieGenreNotFound: // Grouped Not Found Errors
			return http.StatusNotFound
		case CodeMethodNotAllowed:
			return http.StatusMethodNotAllowed
		case CodePaymentRequired, CodeInsufficientFunds, CodeTransactionDeclined: // Grouped Payment Errors
			return http.StatusPaymentRequired
		case CodeRateLimitExceeded:
			return http.StatusTooManyRequests
		case CodeRequestBodyTooLarge:
			return http.StatusRequestEntityTooLarge
		case CodeRequestTimeout: // Client-side request timeout
			return http.StatusRequestTimeout
		case CodeUnsupportedMediaType:
			return http.StatusUnsupportedMediaType
		case CodeUnprocessable, CodeValidation, CodeDataException, CodeForeignKeyViolation,
			CodeInvalidDateFormat, CodeInvalidEmailFormat, CodeInvalidName, CodeInvalidPasswordStrength,
			CodeInvalidPhoneNumber, CodeInvalidUsername, CodeMovieRatingOutOfBounds, CodeBillingInfoMissing: // Grouped Validation/Unprocessable Errors
			return http.StatusUnprocessableEntity // 422 is generally better for semantic validation issues

		// 5xx Server Error Codes
		case CodeInternal, CodeDatabaseConnectionFailed, CodeDatabaseError, CodeDatabaseTimeout,
			CodeDecryptionFailed, CodeDiskSpaceFull, CodeEncryptionFailed, CodeFileError,
			CodeFailedToRetrieveTokenHash, CodeFileSystemReadOnly, CodeFileUploadFailed,
			CodeHashFunctionUnavailable, CodeMissingJWTSecret, CodeNoLogger,
			CodeRedisNotConfigured, CodeResourceCreationFailed, CodeResourceDeletionFailed,
			CodeResourceUpdateFailed, CodeTransactionFailed, CodeTokenHashFailed,
			CodeTokenHashFailedInternal, CodeUnexpectedSystemError, CodeQueryFailed,
			CodeExternalAPIError, CodeExternalAPIRequestFailed, CodeExternalAPIServiceUnavailable,
			CodeNetworkError, CodeNetworkUnreachable, CodeDNSResolutionFailed,
			CodePaymentGatewayTimeout, CodeTimeoutWaitingForResponse, CodeFailedJSONResWrite: // Grouped Internal Server Errors
			return http.StatusInternalServerError
		case CodeNotImplemented, CodeFeatureNotImplemented, CodeUnsupportedOperation: // Grouped Not Implemented Errors
			return http.StatusNotImplemented
		case CodeServiceUnavailable, CodeMaintenanceModeActive: // Grouped Service Unavailable Errors
			return http.StatusServiceUnavailable
		case CodeGatewayTimeout:
			return http.StatusGatewayTimeout

		default:
			// Fallback for unmapped string codes - indicates a missing mapping or new error code
			return http.StatusInternalServerError
		}
	case int:
		// If the code is already an int, assume it's a valid HTTP status code.
		// This allows direct use of http.Status... constants if desired.
		return c
	default:
		// Fallback for unexpected `code` types (e.g., nil, bool, struct)
		return http.StatusInternalServerError
	}
}

// =====================================
// APPLICATION ERROR CODES (STRING CONSTANTS)
// These constants define unique, machine-readable string identifiers for various
// error conditions within the application. They are used for internal error
// identification, logging, and as a 'code' field in structured API error responses.
// They are distinct from HTTP status codes, though they often map to them.
// =====================================
const (
	// --- General & Common Application Codes ---
	CodeAccepted             = "ACCEPTED"               // Success: Operation accepted for processing (e.g., async task)
	CodeBadRequest           = "BAD_REQUEST"            // Client: General malformed request, invalid syntax.
	CodeConflict             = "CONFLICT"               // Client: Request conflicts with current resource state (e.g., duplicate).
	CodeCreated              = "CREATED"                // Success: Resource successfully created.
	CodeForbidden            = "FORBIDDEN"              // Client: Authenticated but unauthorized access.
	CodeGatewayTimeout       = "GATEWAY_TIMEOUT"        // Server: Gateway did not receive timely response from upstream.
	CodeInternal             = "INTERNAL_SERVER_ERROR"  // Server: Generic unexpected server-side error.
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"     // Client: HTTP method not supported for resource.
	CodeNoContent            = "NO_CONTENT"             // Success: Request processed, no content to return.
	CodeNotFound             = "NOT_FOUND"              // Client: Requested resource/entity does not exist.
	CodeNotImplemented       = "NOT_IMPLEMENTED"        // Server: Feature/operation not implemented.
	CodeOK                   = "OK"                     // Success: Request successfully processed.
	CodePaymentRequired      = "PAYMENT_REQUIRED"       // Client: Payment is required.
	CodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"    // Client: Too many requests from the client.
	CodeRequestBodyTooLarge  = "REQUEST_BODY_TOO_LARGE" // Client: Request payload too large.
	CodeServiceUnavailable   = "SERVICE_UNAVAILABLE"    // Server: Service temporarily overloaded or down.
	CodeUnauthorized         = "UNAUTHORIZED"           // Client: Authentication failed or not provided.
	CodeUnprocessable        = "UNPROCESSABLE_ENTITY"   // Client: Semantic errors, validation failure, unprocessable data.
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE" // Client: Request content type not supported.
	CodeValidation           = "VALIDATION_ERROR"       // Client: Generic validation failure.
	CodeRedisURLNotSet       = "REDIS_CONFIG_MISSING"
	CodeValidationError      = "VALIDATION_FAILED"

	// --- Authentication & Authorization Specific Codes ---
	CodeAccountLocked           = "ACCOUNT_LOCKED"
	CodeAccountSuspended        = "ACCOUNT_SUSPENDED"
	CodeAdminOnlyResource       = "ADMIN_ONLY_RESOURCE"
	CodeCSRFTokenMissing        = "CSRF_TOKEN_MISSING"
	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	CodeInvalidAPIKey           = "INVALID_API_KEY"
	CodeInvalidAuth             = "INVALID_AUTH"
	CodeInvalidJWT              = "INVALID_JWT"   // JWT is syntactically invalid or corrupted.
	CodeInvalidToken            = "INVALID_TOKEN" // General invalid token (e.g., revoked, non-existent, unknown reason).
	CodeInvalidTokenSignature   = "INVALID_TOKEN_SIGNATURE"
	CodeMissingJWTSecret        = "JWT_SECRET_MISSING" // Internal: JWT secret not configured.
	CodeMissingToken            = "TOKEN_MISSING"      // Authentication token not provided.
	CodeOAuthError              = "OAUTH_ERROR"
	CodeOAuthTokenExpired       = "OAUTH_TOKEN_EXPIRED"
	CodeSessionExpired          = "SESSION_EXPIRED"
	CodeTokenExpired            = "TOKEN_EXPIRED"
	CodeTokenHashFailed         = "TOKEN_HASH_FAILED" // Internal: Failed to hash a token internally.
	CodeTokenMalformed          = "TOKEN_MALFORMED"
	CodeTokenNotActive          = "TOKEN_NOT_ACTIVE"

	// --- User & Account Management Codes ---
	CodeInvalidDateFormat       = "INVALID_DATE_FORMAT"
	CodeInvalidEmailFormat      = "INVALID_EMAIL_FORMAT"
	CodeInvalidName             = "INVALID_NAME"
	CodeInvalidPasswordStrength = "PASSWORD_STRENGTH_FAILED"
	CodeInvalidPhoneNumber      = "INVALID_PHONE_NUMBER"
	CodeInvalidUsername         = "INVALID_USERNAME"
	CodeMissingUserContext      = "MISSING_USER_CONTEXT" // Internal: User context not found in request.
	CodeRegistrationFailed      = "REGISTRATION_FAILED"
	CodeUserExists              = "USER_EXISTS"
	CodeUserNotFound            = "USER_NOT_FOUND"

	// --- Database & Persistence Codes ---
	CodeDataException            = "DATA_EXCEPTION"             // Data integrity issue (e.g., data too long, invalid type cast).
	CodeDatabaseConnectionFailed = "DATABASE_CONNECTION_FAILED" // Cannot connect to the database.
	CodeDatabaseError            = "DATABASE_ERROR"             // Generic database operation error.
	CodeDatabaseTimeout          = "DATABASE_TIMEOUT"           // Database query or connection timed out.
	CodeDeadlock                 = "DEADLOCK_DETECTED"          // Concurrency deadlock in the database.
	CodeDuplicateEntry           = "DUPLICATE_ENTRY"            // Attempted to insert a duplicate record (e.g., unique constraint).
	CodeForeignKeyViolation      = "FOREIGN_KEY_VIOLATION"      // Database foreign key constraint violated.
	CodeQueryFailed              = "QUERY_FAILED"               // General database query failure.
	CodeResourceCreationFailed   = "RESOURCE_CREATION_FAILED"   // Failed to persist a new resource.
	CodeResourceDeletionFailed   = "RESOURCE_DELETION_FAILED"   // Failed to delete a resource.
	CodeResourceUpdateFailed     = "RESOURCE_UPDATE_FAILED"     // Failed to update an existing resource.
	CodeTransactionFailed        = "TRANSACTION_FAILED"         // A database transaction failed or was rolled back.

	// --- Cryptography & Security Codes ---
	CodeDecryptionFailed          = "DECRYPTION_FAILED"
	CodeEncryptionFailed          = "ENCRYPTION_FAILED"
	CodeFailedToRetrieveTokenHash = "FAILED_TO_RETRIEVE_TOKEN_HASH"
	CodeHashFunctionUnavailable   = "HASH_FUNCTION_UNAVAILABLE"
	CodeInvalidKey                = "INVALID_KEY"
	CodeInvalidKeyType            = "INVALID_KEY_TYPE"
	CodeInvalidSignature          = "INVALID_SIGNATURE"

	// --- File & Storage Codes ---
	CodeDirectoryNotFound   = "DIRECTORY_NOT_FOUND"
	CodeDiskSpaceFull       = "DISK_SPACE_FULL"
	CodeFileError           = "FILE_ERROR" // General file operation error.
	CodeFileNotFound        = "FILE_NOT_FOUND"
	CodeFileSystemReadOnly  = "FILE_SYSTEM_READ_ONLY"
	CodeFileTooLarge        = "FILE_TOO_LARGE"
	CodeFileUploadFailed    = "FILE_UPLOAD_FAILED"
	CodeUnsupportedFileType = "UNSUPPORTED_FILE_TYPE"

	// --- Network & External Service Codes ---
	CodeBillingInfoMissing            = "BILLING_INFO_MISSING"
	CodeDNSResolutionFailed           = "DNS_RESOLUTION_FAILED"
	CodeExternalAPIError              = "EXTERNAL_API_ERROR" // General error from an external API.
	CodeExternalAPIRequestFailed      = "EXTERNAL_API_REQUEST_FAILED"
	CodeExternalAPIServiceUnavailable = "EXTERNAL_API_SERVICE_UNAVAILABLE"
	CodeInsufficientFunds             = "INSUFFICIENT_FUNDS"
	CodeInvalidPaymentMethod          = "INVALID_PAYMENT_METHOD"
	CodeNetworkError                  = "NETWORK_ERROR" // General network connectivity issue.
	CodeNetworkUnreachable            = "NETWORK_UNREACHABLE"
	CodePaymentGatewayTimeout         = "PAYMENT_GATEWAY_TIMEOUT"
	CodeTransactionDeclined           = "TRANSACTION_DECLINED"
	CodeTimeoutWaitingForResponse     = "TIMEOUT_WAITING_FOR_RESPONSE" // General network timeout.

	// --- System, Configuration & Maintenance Codes ---
	CodeFailedJSONResWrite      = "FAILED_JSON_RESPONSE_WRITE"
	CodeFeatureNotImplemented   = "FEATURE_NOT_IMPLEMENTED"
	CodeMaintenanceModeActive   = "MAINTENANCE_MODE_ACTIVE"
	CodeNoLogger                = "NO_LOGGER" // Internal: Logger instance is missing.
	CodeRedisNotConfigured      = "REDIS_NOT_CONFIGURED"
	CodeRequestTimeout          = "REQUEST_TIMEOUT"            // Server timed out waiting for the client request.
	CodeTokenHashFailedInternal = "TOKEN_HASH_FAILED_INTERNAL" // Internal: Token hash operation failed.
	CodeUnexpectedSystemError   = "UNEXPECTED_SYSTEM_ERROR"
	CodeUnsupportedLanguage     = "UNSUPPORTED_LANGUAGE"
	CodeUnsupportedOperation    = "UNSUPPORTED_OPERATION"

	// --- Domain-Specific Codes (e.g., Movie, Actor, Genre) ---
	CodeInvalidActorID         = "INVALID_ACTOR_ID"
	CodeInvalidGenreID         = "INVALID_GENRE_ID"
	CodeInvalidMovieID         = "INVALID_MOVIE_ID"
	CodeMovieAlreadyExists     = "MOVIE_ALREADY_EXISTS"
	CodeMovieCastNotFound      = "MOVIE_CAST_NOT_FOUND"
	CodeMovieCreateFailed      = "MOVIE_CREATE_FAILED"
	CodeMovieDeleteFailed      = "MOVIE_DELETE_FAILED"
	CodeMovieGenreNotFound     = "MOVIE_GENRE_NOT_FOUND"
	CodeMovieNotFound          = "MOVIE_NOT_FOUND"
	CodeMovieRatingOutOfBounds = "MOVIE_RATING_OUT_OF_BOUNDS"
	CodeMovieReviewFailed      = "MOVIE_REVIEW_FAILED"
	CodeMovieUpdateFailed      = "MOVIE_UPDATE_FAILED"
)

// HTTPStatus returns the HTTP status code for a given app error code.
// func HTTPStatus(code any) int {
// 	switch c := code.(type) {
// 	case string:
// 		switch c {
// 		// 2xx Success
// 		case CodeOK:
// 			return http.StatusOK
// 		case CodeAccepted:
// 			return http.StatusAccepted
// 		case CodeCreated:
// 			return http.StatusCreated
// 		case CodeNoContent:
// 			return http.StatusNoContent

// 		// 4xx Client Errors
// 		case CodeAccountLocked, CodeAccountSuspended, CodeInvalidAuth, CodeInvalidAPIKey,
// 			CodeInvalidJWT, CodeOAuthError, CodeOAuthTokenExpired, CodeSessionExpired,
// 			CodeTokenExpired, CodeTokenInvalid, CodeTokenMalformed, CodeTokenMissing:
// 			return http.StatusUnauthorized // Authentication errors

// 		case CodeAdminOnlyResource, CodeCSRFTokenMissing, CodeForbidden, CodeInsufficientPermissions:
// 			return http.StatusForbidden // Authorization errors

// 		case CodeBadRequest, CodeFieldRequired, CodeFileUploadFailed, CodeFileTooLarge,
// 			CodeInvalidContentType, CodeInvalidDateFormat, CodeInvalidEmailFormat,
// 			CodeInvalidIDParameter, CodeInvalidMovieID, CodeInvalidActorID, CodeInvalidGenreID,
// 			CodeInvalidRequestPayload, CodePasswordStrength, CodeRegFailure,
// 			CodeUnsupportedFileType:
// 			return http.StatusBadRequest // General client input/request errors

// 		case CodeConflict, CodeDuplicateEntry, CodeUserExists:
// 			return http.StatusConflict // Resource conflicts

// 		case CodeMethodNotAllowed:
// 			return http.StatusMethodNotAllowed

// 		case CodeFileNotFound, CodeDirectoryNotFound, CodeNotFound, CodeUserNotFound:
// 			return http.StatusNotFound // Resource not found

// 		case CodePaymentRequired:
// 			return http.StatusPaymentRequired

// 		case CodeRateLimitExceeded:
// 			return http.StatusTooManyRequests

// 		case CodeRequestBodyTooLarge:
// 			return http.StatusRequestEntityTooLarge

// 		case CodeUnsupportedMediaType:
// 			return http.StatusUnsupportedMediaType

// 		case CodeDataException, CodeForeignKeyViolation, CodeUnprocessable, CodeValidationError:
// 			return http.StatusUnprocessableEntity // Semantic errors, validation, data integrity

// 		// 5xx Server Errors
// 		case CodeDeadlock, CodeExternalAPIError, CodeExternalAPIRequestFailed,
// 			CodeExternalAPIServiceUnavailable, CodeGatewayTimeout, CodeMaintenanceModeActive,
// 			CodeNetworkError, CodeNetworkUnreachable, CodeDNSResolutionFailed,
// 			CodeServiceUnavailable, CodeTimeoutWaitingForResponse:
// 			return http.StatusServiceUnavailable // Transient server issues or external dependencies

// 		case CodeDatabaseConnectionFailed, CodeDatabaseError, CodeDatabaseTimeout,
// 			CodeDecryptionFailed, CodeDiskSpaceFull, CodeEncryptionFailed, CodeFileError,
// 			CodeFailedToRetrieveTokenHash, CodeFeatureNotImplemented, CodeFileSystemReadOnly,
// 			CodeHashFunctionUnavailable, CodeInternal, CodeInvalidKey, CodeInvalidKeyType,
// 			CodeMissingJWTSecret, CodeMissingUserContext, CodeNotImplemented,
// 			CodeRedisURLNotSet, CodeRequestTimeout, CodeResourceCreationFailed,
// 			CodeResourceDeletionFailed, CodeResourceUpdateFailed, CodeTransactionFailed,
// 			CodeUnexpectedSystemError, CodeUnsupportedLanguage, CodeUnsupportedOperation:
// 			return http.StatusInternalServerError // Internal server errors

// 		default:
// 			// Fallback for unmapped string codes
// 			return http.StatusInternalServerError
// 		}
// 	case int:
// 		// Directly map the int codes to their HTTP status values
// 		return c
// 	default:
// 		// Fallback for unexpected `code` types or unmapped codes
// 		return http.StatusInternalServerError
// 	}
// }

// =====================================
// HTTP Status Codes for Application Errors
// These constants define the HTTP status codes returned by the application
// for various error and success conditions. They directly map to Go's standard
// library net/http status constants for consistency and clarity.
// =====================================
// const (
// 	// --- 2xx Success Codes ---
// 	CodeOK        = http.StatusOK        // 200 OK: The request was successful.
// 	CodeCreated   = http.StatusCreated   // 201 Created: The request has been fulfilled and resulted in a new resource being created.
// 	CodeAccepted  = http.StatusAccepted  // 202 Accepted: The request has been accepted for processing.
// 	CodeNoContent = http.StatusNoContent // 204 No Content: The server successfully processed the request and is not returning any content.

// 	// --- 3xx Redirection Codes ---
// 	// No specific 3xx codes were explicitly listed, but they can be added here if needed.

// 	// --- 4xx Client Error Codes ---
// 	CodeBadRequest           = http.StatusBadRequest            // 400 Bad Request: The server cannot or will not process the request due to an apparent client error.
// 	CodeUnauthorized         = http.StatusUnauthorized          // 401 Unauthorized: Authentication is required or has failed.
// 	CodePaymentRequired      = http.StatusPaymentRequired       // 402 Payment Required: Payment is required to access this resource.
// 	CodeForbidden            = http.StatusForbidden             // 403 Forbidden: The client does not have access rights to the content.
// 	CodeNotFound             = http.StatusNotFound              // 404 Not Found: The requested resource could not be found.
// 	CodeMethodNotAllowed     = http.StatusMethodNotAllowed      // 405 Method Not Allowed: The request method is not supported for the requested resource.
// 	CodeRequestTimeout       = http.StatusRequestTimeout        // 408 Request Timeout: The server timed out waiting for the request.
// 	CodeConflict             = http.StatusConflict              // 409 Conflict: The request could not be completed due to a conflict with the current state of the resource (e.g., duplicate entry).
// 	CodeRequestBodyTooLarge  = http.StatusRequestEntityTooLarge // 413 Payload Too Large: The request payload is larger than the server can process.
// 	CodeUnsupportedMediaType = http.StatusUnsupportedMediaType  // 415 Unsupported Media Type: The media format of the requested data is not supported by the server.
// 	CodeUnprocessable        = http.StatusUnprocessableEntity   // 422 Unprocessable Content: The server understands the content type, but was unable to process the contained instructions (e.g., validation failure).
// 	CodeTooManyRequests      = http.StatusTooManyRequests       // 429 Too Many Requests: The user has sent too many requests in a given amount of time (rate limiting).

// 	// --- 5xx Server Error Codes ---
// 	CodeInternal           = http.StatusInternalServerError // 500 Internal Server Error: A generic error message, given when an unexpected condition was encountered.
// 	CodeNotImplemented     = http.StatusNotImplemented      // 501 Not Implemented: The server does not support the functionality required to fulfill the request.
// 	CodeServiceUnavailable = http.StatusServiceUnavailable  // 503 Service Unavailable: The server is not ready to handle the request (e.g., overloaded or down for maintenance).
// 	CodeGatewayTimeout     = http.StatusGatewayTimeout      // 504 Gateway Timeout: The server, while acting as a gateway, did not get a response in time.
// )

// const (
// 	// =====================================
// 	// GENERAL HTTP STATUS-ALIGNED CODES
// 	// =====================================
// 	CodeAccepted             = "ACCEPTED"               // 202 Accepted
// 	CodeBadRequest           = "BAD_REQUEST"            // 400 Bad Request: General client-side input error, malformed syntax.
// 	CodeConflict             = "CONFLICT"               // 409 Conflict: Request conflicts with current state (e.g., duplicate entry, optimistic locking failure).
// 	CodeCreated              = "CREATED"                // 201 Created
// 	CodeForbidden            = "FORBIDDEN"              // 403 Forbidden: Authenticated but not authorized to perform action.
// 	CodeGatewayTimeout       = "GATEWAY_TIMEOUT"        // 504 Gateway Timeout: The server, while acting as a gateway, did not get a response in time.
// 	CodeInternal             = "INTERNAL_SERVER_ERROR"  // 500 Internal Server Error: Generic server-side error not covered by more specific categories.
// 	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"     // 405 Method Not Allowed: HTTP method not supported for the resource.
// 	CodeNoContent            = "NO_CONTENT"             // 204 No Content
// 	CodeNotFound             = "NOT_FOUND"              // 404 Not Found: Requested resource does not exist.
// 	CodeNotImplemented       = "NOT_IMPLEMENTED"        // 501 Not Implemented
// 	CodeOK                   = "OK"                     // 200 OK
// 	CodePaymentRequired      = "PAYMENT_REQUIRED"       // 402 Payment Required
// 	CodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"    // 429 Too Many Requests
// 	CodeRequestBodyTooLarge  = "REQUEST_BODY_TOO_LARGE" // 413 Payload Too Large
// 	CodeServiceUnavailable   = "SERVICE_UNAVAILABLE"    // 503 Service Unavailable: The server is temporarily unable to handle the request.
// 	CodeUnauthorized         = "UNAUTHORIZED"           // 401 Unauthorized: Authentication failed or not provided.
// 	CodeUnprocessable        = "UNPROCESSABLE_ENTITY"   // 422 Unprocessable Content: Semantic errors in request, validation failure, or unprocessable data.
// 	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE" // 415 Unsupported Media Type

// 	// =====================================
// 	// AUTHENTICATION & AUTHORIZATION SPECIFIC ERRORS
// 	// Errors related to identity and access management.
// 	// =====================================
// 	CodeAccountLocked           = "ACCOUNT_LOCKED"           // User account is locked.
// 	CodeAccountSuspended        = "ACCOUNT_SUSPENDED"        // User account is suspended.
// 	CodeAdminOnlyResource       = "ADMIN_ONLY_RESOURCE"      // Attempted access to admin-only resource by non-admin.
// 	CodeCSRFTokenMissing        = "CSRF_TOKEN_MISSING"       // CSRF token not found in request.
// 	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS" // User lacks necessary permissions.
// 	CodeInvalidAuth             = "INVALID_AUTH"             // General invalid authentication details.
// 	CodeInvalidAPIKey           = "INVALID_API_KEY"          // The provided API key is invalid or unauthorized.
// 	CodeInvalidJWT              = "INVALID_JWT"              // JWT is syntactically invalid or corrupted.
// 	CodeMissingJWTSecret        = "JWT_SECRET_MISSING"       // Internal server error: JWT secret not configured.
// 	CodeOAuthError              = "OAUTH_ERROR"              // General OAuth related error.
// 	CodeOAuthTokenExpired       = "OAUTH_TOKEN_EXPIRED"      // OAuth token has expired.
// 	CodeSessionExpired          = "SESSION_EXPIRED"          // User session is no longer active.
// 	CodeTokenExpired            = "TOKEN_EXPIRED"            // Authentication token has expired.
// 	CodeTokenHashFailed         = "TOKEN_HASH_FAILED"        // Failed to hash a token internally.
// 	CodeTokenInvalid            = "TOKEN_INVALID"            // Token is valid but structurally incorrect or revoked.
// 	CodeTokenMalformed          = "TOKEN_MALFORMED"          // Token format is incorrect.
// 	CodeTokenMissing            = "TOKEN_MISSING"            // Authentication token not provided.

// 	// =====================================
// 	// USER & ACCOUNT MANAGEMENT ERRORS
// 	// Errors specific to user lifecycle and account states.
// 	// =====================================
// 	CodeMissingUserContext = "MISSING_USER_CONTEXT" // Internal error: user context not found in request.
// 	CodeRegFailure         = "REGISTRATION_FAILED"  // General user registration failure.
// 	CodeUserExists         = "USER_EXISTS"          // Attempt to create a user that already exists.
// 	CodeUserNotFound       = "USER_NOT_FOUND"       // Specific: user not found. (overlaps with CodeNotFound but for explicit user context)

// 	// =====================================
// 	// REQUEST & VALIDATION SPECIFIC ERRORS
// 	// Errors for invalid client requests or data.
// 	// =====================================
// 	CodeFieldRequired         = "FIELD_REQUIRED"            // A mandatory field is missing.
// 	CodeInvalidContentType    = "INVALID_CONTENT_TYPE"      // Request Content-Type header is not supported.
// 	CodeInvalidDateFormat     = "INVALID_DATE_FORMAT"       // Date string provided is in an unrecognized format.
// 	CodeInvalidEmailFormat    = "INVALID_EMAIL_FORMAT"      // Provided email does not meet format requirements.
// 	CodeInvalidIDParameter    = "INVALID_ID_PARAMETER"      // ID in path/query parameter is invalid format (e.g., non-UUID).
// 	CodeInvalidRequestPayload = "INVALID_REQUEST_PAYLOAD"   // JSON/XML payload is unparseable or malformed.
// 	CodePasswordStrength      = "PASSWORD_STRENGTH_FAILED"  // Password does not meet strength requirements.
// 	CodeValidationError       = "VALIDATION_ERROR"          // Generic validation failure (use metadata for details).
// 	CodeRequestEntityTooLarge = "REQUEST_ENTITY_TOO_LARGE " // http.StatusRequestEntityTooLarge is set to 413
// 	// =====================================
// 	// RESOURCE & DOMAIN-SPECIFIC IDENTIFIER ERRORS
// 	// Errors related to specific resource IDs being invalid.
// 	// =====================================
// 	CodeInvalidActorID = "INVALID_ACTOR_ID" // Specific to actor resource ID.
// 	CodeInvalidGenreID = "INVALID_GENRE_ID" // Specific to genre resource ID.
// 	CodeInvalidMovieID = "INVALID_MOVIE_ID" // Specific to movie resource ID.

// 	// =====================================
// 	// DATABASE & PERSISTENCE ERRORS
// 	// Errors stemming from database operations.
// 	// =====================================
// 	CodeDataException            = "DATA_EXCEPTION"             // Data integrity issue (e.g., data too long, invalid type cast).
// 	CodeDatabaseConnectionFailed = "DATABASE_CONNECTION_FAILED" // Cannot connect to the database.
// 	CodeDatabaseError            = "DATABASE_ERROR"             // Generic database operation error.
// 	CodeDatabaseTimeout          = "DATABASE_TIMEOUT"           // Database query or connection timed out.
// 	CodeDeadlock                 = "DEADLOCK_DETECTED"          // Concurrency deadlock in the database.
// 	CodeDuplicateEntry           = "DUPLICATE_ENTRY"            // Attempted to insert a duplicate record (e.g., unique constraint).
// 	CodeForeignKeyViolation      = "FOREIGN_KEY_VIOLATION"      // Database foreign key constraint violated.
// 	CodeResourceCreationFailed   = "RESOURCE_CREATION_FAILED"   // Failed to persist a new resource.
// 	CodeResourceDeletionFailed   = "RESOURCE_DELETION_FAILED"   // Failed to delete a resource.
// 	CodeResourceUpdateFailed     = "RESOURCE_UPDATE_FAILED"     // Failed to update an existing resource.
// 	CodeTransactionFailed        = "TRANSACTION_FAILED"         // A database transaction failed or was rolled back.
// 	CodeTooManyRequests          = "TOO_MANY_REQUESTS"          // http.StatusTooManyRequests is set to 429
// 	// =====================================
// 	// CRYPTOGRAPHY & SECURITY ERRORS
// 	// Errors related to encryption, hashing, and token security.
// 	// =====================================
// 	CodeDecryptionFailed          = "DECRYPTION_FAILED"             // Data decryption failed.
// 	CodeEncryptionFailed          = "ENCRYPTION_FAILED"             // Data encryption failed.
// 	CodeFailedToRetrieveTokenHash = "FAILED_TO_RETRIEVE_TOKEN_HASH" // Could not retrieve token hash for comparison.
// 	CodeHashFunctionUnavailable   = "HASH_FUNCTION_UNAVAILABLE"     // Required hash function is not available.
// 	CodeInvalidKey                = "INVALID_KEY"                   // Cryptographic key is invalid.
// 	CodeInvalidKeyType            = "INVALID_KEY_TYPE"              // Cryptographic key type is unsupported.
// 	CodeInvalidSignature          = "INVALID_SIGNATURE"             // Digital signature is invalid.

// 	// =====================================
// 	// FILE & STORAGE ERRORS
// 	// Errors related to file system operations and storage.
// 	// =====================================
// 	CodeDirectoryNotFound   = "DIRECTORY_NOT_FOUND"   // Specified directory does not exist.
// 	CodeFileError           = "FILE_ERROR"            // General file operation error.
// 	CodeFileNotFound        = "FILE_NOT_FOUND"        // Requested file does not exist.
// 	CodeFileSystemReadOnly  = "FILE_SYSTEM_READ_ONLY" // File system is in read-only mode.
// 	CodeFileUploadFailed    = "FILE_UPLOAD_FAILED"    // Failed to upload a file.
// 	CodeFileTooLarge        = "FILE_TOO_LARGE"        // File size exceeds allowed limit.
// 	CodeDiskSpaceFull       = "DISK_SPACE_FULL"       // Storage disk space is exhausted.
// 	CodeUnsupportedFileType = "UNSUPPORTED_FILE_TYPE" // Attempted to upload/process an unsupported file type.

// 	// =====================================
// 	// NETWORK & EXTERNAL SERVICE ERRORS
// 	// Errors related to network connectivity or external API interactions.
// 	// =====================================
// 	CodeDNSResolutionFailed           = "DNS_RESOLUTION_FAILED"            // DNS lookup failed.
// 	CodeExternalAPIError              = "EXTERNAL_API_ERROR"               // General error from an external API.
// 	CodeExternalAPIRequestFailed      = "EXTERNAL_API_REQUEST_FAILED"      // Request to external API failed for unknown reason.
// 	CodeExternalAPIServiceUnavailable = "EXTERNAL_API_SERVICE_UNAVAILABLE" // External API is down or unreachable.
// 	CodeNetworkError                  = "NETWORK_ERROR"                    // General network connectivity issue.
// 	CodeNetworkUnreachable            = "NETWORK_UNREACHABLE"              // Network destination is unreachable.
// 	CodeTimeoutWaitingForResponse     = "TIMEOUT_WAITING_FOR_RESPONSE"     // General network timeout.

// 	// =====================================
// 	// SYSTEM, CONFIGURATION & MAINTENANCE ERRORS
// 	// Errors indicating internal system issues or operational states.
// 	// =====================================
// 	CodeFeatureNotImplemented = "FEATURE_NOT_IMPLEMENTED" // Requested feature is not yet implemented.
// 	CodeMaintenanceModeActive = "MAINTENANCE_MODE_ACTIVE" // System is in maintenance mode.
// 	CodeNoLoggerErr           = "NO_LOGGER_ERROR"         // Internal: Logger instance is missing.
// 	CodeRedisURLNotSet        = "REDIS_URL_NOT_SET"       // Internal: Redis URL configuration is missing.
// 	CodeRequestTimeout        = "REQUEST_TIMEOUT"         // The server timed out waiting for the request.
// 	CodeUnexpectedSystemError = "UNEXPECTED_SYSTEM_ERROR" // An unforeseen system error occurred.
// 	CodeUnsupportedLanguage   = "UNSUPPORTED_LANGUAGE"    // Requested language is not supported.
// 	CodeUnsupportedOperation  = "UNSUPPORTED_OPERATION"   // Operation is not supported (e.g., for resource type).
// )

// const (
// 	CodeUnauthorized          = http.StatusUnauthorized          // 401
// 	CodeForbidden             = http.StatusForbidden             // 403
// 	CodeNotFound              = http.StatusNotFound              // 404
// 	CodeConflict              = http.StatusConflict              // 409
// 	CodeInternal              = http.StatusInternalServerError   // 500
// 	CodeBadRequest            = http.StatusBadRequest            // 400
// 	CodeUnprocessable         = http.StatusUnprocessableEntity   // 422 (for validation errors)
// 	CodeServiceUnavailable    = http.StatusServiceUnavailable    // 503
// 	CodeTooManyRequests       = http.StatusTooManyRequests       // 429
// 	CodePaymentRequired       = http.StatusPaymentRequired       // 402
// 	CodeUnsupportedMediaType  = http.StatusUnsupportedMediaType  // 415
// 	CodeRequestEntityTooLarge = http.StatusRequestEntityTooLarge // 413
// 	CodeRequestTimeout        = http.StatusRequestTimeout        // 408
// 	CodeNotImplemented        = http.StatusNotImplemented        // 501
// 	CodeDatabaseError         = http.StatusInternalServerError
// )

// const (
// 	// =============================
// 	// COMMON ERRORS
// 	// =============================
// 	CodeBadRequest    = "bad_request"
// 	CodeNotFound      = "not_found"
// 	CodeConflict      = "conflict"
// 	CodeUnprocessable = "unprocessable_entity"
// 	CodeInternal      = "internal_server_error"

// 	// =============================
// 	// AUTHENTICATION & AUTHORIZATION ERRORS
// 	// =============================
// 	CodeUnauthorized      = "unauthorized"
// 	CodeForbidden         = "forbidden"
// 	CodeInvalidAuth       = "invalid_auth"
// 	CodeInvalidJWT        = "invalid_jwt"
// 	CodeTokenExpired      = "token_expired"
// 	CodeTokenInvalid      = "token_invalid"
// 	CodeTokenMissing      = "token_missing"
// 	CodeTokenMalformed    = "token_malformed"
// 	CodeMissingJWTSecret  = "jwt_secret_missing"
// 	CodeOAuthTokenExpired = "oauth_token_expired"
// 	CodeOAuthError        = "oauth_error"
// 	CodeTokenHashFailed   = "token_hash_failed" // Fixed constant casing to snake_case

// 	// =============================
// 	// USER MANAGEMENT ERRORS
// 	// =============================
// 	CodeUserNotFound            = "user_not_found"
// 	CodeUserExists              = "user_exists"
// 	CodeRegFailure              = "registration_failed"
// 	CodeMissingUserContext      = "missing_user_context"
// 	CodeAdminOnlyResource       = "admin_only_resource"
// 	CodeAccountLocked           = "account_locked"
// 	CodeAccountSuspended        = "account_suspended"
// 	CodeSessionExpired          = "session_expired"
// 	CodeInsufficientPermissions = "insufficient_permissions"

// 	// =============================
// 	// RESOURCE ERRORS
// 	// =============================
// 	CodeResourceNotFound = "resource_not_found"
// 	CodeRecordNotFound   = "record_not_found"
// 	CodeInvalidMovieID   = "invalid_movie_id"
// 	CodeInvalidActorID   = "invalid_actor_id"
// 	CodeInvalidGenreID   = "invalid_genre_id"

// 	// =============================
// 	// REQUEST & VALIDATION ERRORS
// 	// =============================
// 	CodeInvalidRequestPayload = "invalid_request_payload"
// 	CodeInvalidRequest        = "invalid_request"
// 	CodeInvalidIDParameter    = "invalid_id_parameter"
// 	CodeInvalidContentType    = "invalid_content_type"
// 	CodeInvalidEmailFormat    = "invalid_email_format"
// 	CodePasswordStrength      = "password_strength"
// 	CodeInvalidDateFormat     = "invalid_date_format"
// 	CodeFieldRequired         = "field_required"
// 	CodeInvalidAPIKey         = "invalid_api_key"
// 	CodeInvalidInput          = "invalid_input"

// 	// =============================
// 	// DATABASE & SERVER ERRORS
// 	// =============================
// 	CodeResourceCreationFailed   = "resource_creation_failed"
// 	CodeResourceUpdateFailed     = "resource_update_failed"
// 	CodeResourceDeletionFailed   = "resource_deletion_failed"
// 	CodeDatabaseConnectionFailed = "database_connection_failed"
// 	CodeQueryTimeout             = "query_timeout"
// 	CodeDuplicateEntry           = "duplicate_entry"
// 	CodeForeignKeyViolation      = "foreign_key_violation"
// 	CodeTransactionFailed        = "transaction_failed"
// 	CodeDatabaseError            = "database_error"
// 	CodeDatabaseTimeout          = "database_timeout"

// 	// =============================
// 	// CRYPTOGRAPHY & KEY MANAGEMENT ERRORS
// 	// =============================
// 	CodeInvalidKey                = "invalid_key"
// 	CodeInvalidKeyType            = "invalid_key_type"
// 	CodeHashFunctionUnavailable   = "hash_function_unavailable"
// 	CodeFailedToRetrieveTokenHash = "failed_to_retrieve_token_hash"
// 	CodeEncryptionFailed          = "encryption_failed"
// 	CodeDecryptionFailed          = "decryption_failed"
// 	CodeInvalidSignature          = "invalid_signature"
// 	CodeCSRFTokenMissing          = "csrf_token_missing"

// 	// =============================
// 	// FILE & STORAGE ERRORS
// 	// =============================
// 	CodeFileNotFound        = "file_not_found"
// 	CodeFileUploadFailed    = "file_upload_failed"
// 	CodeUnsupportedFileType = "unsupported_file_type"
// 	CodeFileTooLarge        = "file_too_large"
// 	CodeDiskSpaceFull       = "disk_space_full"
// 	CodeFileSystemReadOnly  = "file_system_read_only"
// 	CodeDirectoryNotFound   = "directory_not_found"
// 	CodeFileError           = "file_error"

// 	// =============================
// 	// NETWORK & EXTERNAL SERVICE ERRORS
// 	// =============================
// 	CodeTimeoutWaitingForResponse     = "timeout_waiting_for_response"
// 	CodeDNSResolutionFailed           = "dns_resolution_failed"
// 	CodeServiceUnavailable            = "service_unavailable"
// 	CodeNetworkUnreachable            = "network_unreachable"
// 	CodeExternalAPIServiceUnavailable = "external_api_service_unavailable"
// 	CodeExternalAPIRequestFailed      = "external_api_request_failed"
// 	CodeRateLimitExceeded             = "rate_limit_exceeded"
// 	CodeNetworkError                  = "network_error"
// 	CodeExternalAPIError              = "external_api_error"

// 	// =============================
// 	// RATE LIMITING ERRORS
// 	// =============================
// 	CodeRateLimitError = "rate_limit_error"

// 	// =============================
// 	// SYSTEM & MAINTENANCE ERRORS
// 	// =============================
// 	CodeUnrecognizedError     = "unrecognized_error"
// 	CodeUnexpectedSystemError = "unexpected_system_error"
// 	CodeMaintenanceModeActive = "maintenance_mode_active"
// 	CodeFeatureNotImplemented = "feature_not_implemented"
// 	CodeUnsupportedOperation  = "unsupported_operation"
// 	CodeUnsupportedLanguage   = "unsupported_language"

// 	// =============================
// 	// GENERAL & CONFIGURATION ERRORS
// 	// =============================
// 	CodeNoLoggerErr         = "no_logger_error"
// 	CodeRedisURLNotSet      = "redis_url_not_set"
// 	CodeRequestTimeout      = "request_timeout"
// 	CodeMethodNotAllowed    = "method_not_allowed"
// 	CodeRequestBodyTooLarge = "request_body_too_large"

// 	// --- Generic HTTP Status-aligned Codes ---
// 	CodeOK               = "OK"
// 	CodeCreated          = "CREATED"
// 	CodeAccepted         = "ACCEPTED"
// 	CodeNoContent        = "NO_CONTENT"
// 	CodeBadRequest       = "BAD_REQUEST"  // General client-side input error
// 	CodeUnauthorized     = "UNAUTHORIZED" // Authentication failed or not provided
// 	CodeForbidden        = "FORBIDDEN"    // Authenticated but not authorized
// 	CodeNotFound         = "NOT_FOUND"    // Resource not found
// 	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
// 	CodeConflict         = "CONFLICT"             // Resource conflict (e.g., duplicate, optimistic lock)
// 	CodeUnprocessable    = "UNPROCESSABLE_ENTITY" // Semantic errors in request, e.g., validation failure

// 	// --- Server-side / Internal Error Codes ---
// 	CodeInternal           = "INTERNAL_SERVER_ERROR" // Generic server-side error
// 	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"   // Temporary server issue
// 	CodeGatewayTimeout     = "GATEWAY_TIMEOUT"       // Upstream service timeout

// 	// --- Specific Custom Error Codes (as requested) ---
// 	CodeValidationError = "VALIDATION_ERROR"  // More specific than CodeUnprocessable, focuses on data validity
// 	CodeDataException   = "DATA_EXCEPTION"    // Generic issue with data integrity or type mismatch in DB/system
// 	CodeDeadlock        = "DEADLOCK_DETECTED" // Database deadlock or similar concurrency conflict
// 	CodeDatabaseError   = "DATABASE_ERROR"    // Generic database operation error not covered by more specific DB codes

// 	// --- Other potential specific codes ---
// 	CodeForeignKeyViolation = "FOREIGN_KEY_VIOLATION"
// )

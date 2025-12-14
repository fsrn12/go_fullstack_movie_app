package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"multipass/pkg/common"
	"multipass/pkg/logging"
	"multipass/pkg/response"
)

// AppError represents a structured application error with detailed context.
type AppError struct {
	Code     any            `json:"code"`
	Message  string         `json:"message"`
	Location string         `json:"location,omitempty"`
	Op       string         `json:"operation,omitempty"`
	Err      error          `json:"-"`
	Stack    string         `json:"stack,omitempty"`
	Metadata common.Envelop `json:"metadata,omitempty"`
	Logger   logging.Logger `json:"-"`
}

// NewAppError creates a new AppError instance.
// It captures the file, line, and function where it's called from.
func NewAppError(code any, message, op string, err error, logger logging.Logger, metadata common.Envelop) *AppError {
	_, file, line, _ := runtime.Caller(1)
	loc := callerLocation(2) // skip 2 frames: this func + caller

	if metadata == nil { // Initialize metadata if nil
		metadata = common.Envelop{}
	}

	return &AppError{
		Code:     code,
		Message:  message,
		Op:       op,
		Location: loc,
		Err:      err,
		Stack:    fmt.Sprintf("%s:%d", file, line),
		Metadata: metadata,
		Logger:   logger,
	}
}

// NewAppErrorf creates a new AppError instance with a formatted message.
func NewAppErrorf(code any, messageFmt string, underlying error, logger logging.Logger, metadata common.Envelop) *AppError {
	_, file, line, _ := runtime.Caller(1)
	loc := callerLocation(2)

	if metadata == nil { // Initialize metadata if nil
		metadata = common.Envelop{}
	}

	return &AppError{
		Code:     code,
		Message:  fmt.Sprintf(messageFmt, metadata), // This will print `metadata` as a single value.
		Location: loc,
		Err:      underlying,
		Stack:    fmt.Sprintf("%s:%d", file, line),
		Metadata: metadata,
		Logger:   logger,
	}
}

// Helper to generate the error with just metadata
func NewErrorWithMetadata(code any, message string, logger logging.Logger, metadata common.Envelop) *AppError {
	return NewAppError(code, message, "no_operation", nil, logger, metadata)
}

// Implements error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (cause: %+v)", e.Code, e.Location, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Location, e.Message)
}

// Allows errors.Is and errors.As to work properly with wrapped errors.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WrapWithAppError wraps an existing error inside a new AppError.
// It also captures the file, line, and function where WrapWithAppError is called.
func WrapWithAppError(code any, message string, underlying error, logger logging.Logger, metadata common.Envelop) *AppError {
	// The `location` parameter was redundant as `callerLocation` already calculates it.
	// We ensure `op` is set for the new error, and if the underlying error is an AppError,
	// we might inherit its op or compose it.
	_, file, line, _ := runtime.Caller(1)
	loc := callerLocation(2)

	if metadata == nil { // Initialize metadata if nil
		metadata = common.Envelop{}
	}

	return &AppError{
		Code:     code,
		Message:  message,
		Location: loc,
		Err:      underlying, // The original error is explicitly wrapped
		Stack:    fmt.Sprintf("%s:%d", file, line),
		Metadata: metadata,
		Logger:   logger,
	}
}

// -------------------------
// Helper: Capture Caller Location
// -------------------------
// callerLocation returns a string representing the file, line, and function name
// of the caller at the specified skip level.
func callerLocation(skip int) string {
	// skip=0 is this function, skip=1 is caller of callerLocation, etc.
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// Get function name
	fn := runtime.FuncForPC(pc)
	fnName := "unknown"
	if fn != nil {
		fnName = fn.Name()
		// Trim to package.func style (last 2 components)
		// For example, "main.MyFunc" or "myproject/mypackage.AnotherFunc"
		parts := strings.Split(fnName, "/")
		if len(parts) > 1 {
			fnName = parts[len(parts)-1]
		}
	}
	return fmt.Sprintf("%s:%d %s", file, line, fnName)
}

// LogError logs the AppError details using the embedded logger.
func (e *AppError) LogError() {
	if e.Logger == nil {
		// Fallback: If logger is nil, print to stderr to ensure error is not swallowed
		fmt.Fprintf(os.Stderr, "AppError: No logger available. Code: %v, Message: %s, Operation: %s, Location: %s, Stack: %s, Metadata: %+v, Cause: %+v\n",
			e.Code, e.Message, e.Op, e.Location, e.Stack, e.Metadata, e.Err)
		return
	}

	// Create a structured log entry. Using %+w for wrapped error if Logger supports it.
	// If the logger expects key-value pairs, adapt this.
	logFields := []interface{}{
		"code", e.Code,
		"message", e.Message,
		"operation", e.Op,
		"location", e.Location,
		"stack", e.Stack,
		"metadata", e.Metadata,
	}
	if e.Err != nil {
		logFields = append(logFields, "cause", e.Err)
	}

	e.Logger.Errorf("AppError occurred: %+w", e.Err, logFields...)
}

// ToErrorResponse converts the AppError into a client-friendly map for JSON response.
func (e *AppError) ToErrorResponse() common.Envelop {
	// Decide what to expose to the client.
	// `Stack` and `Location` are often internal.
	response := common.Envelop{
		"error":   true,
		"code":    e.Code,
		"message": e.Message,
	}
	if len(e.Metadata) > 0 {
		response["meta"] = e.Metadata
	}
	return response
}

func (e *AppError) WriteJSONError(w http.ResponseWriter, r *http.Request, jw response.Writer) {
	var statusCode int
	if cStr, ok := e.Code.(string); ok {
		statusCode = HTTPStatus(cStr)
	} else if cInt, ok := e.Code.(int); ok {
		statusCode = cInt
	} else {
		// Fallback if Code is neither a string nor an int (shouldn't happen with your design)
		statusCode = http.StatusInternalServerError
	}

	if e.Logger != nil {
		// Log the outgoing response for debugging/auditing
		e.Logger.Info("Writing error response", "response_payload", e.ToErrorResponse(), "http_status", statusCode)
	}

	// If JSONWriter is nil
	if jw == nil {
		// This indicates a critical setup error. Log it and attempt a basic JSON write.
		if e.Logger != nil {
			e.Logger.Error("Failed to get JSONWriter (jw) instance for writing error response", nil)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Fallback status
		fmt.Fprintf(w, `{"error":true,"code":"INTERNAL_SERVER_ERROR","message":"Internal server error: JSON writer not available."}`)
		return
	}

	err := jw.WriteJSON(w, statusCode, e.ToErrorResponse())
	if err != nil {
		if e.Logger != nil {
			e.Logger.Error("Failed to write JSON response after error", err)
		}
		// Fallback if writing JSON itself fails (rare but important)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":true,"code":"INTERNAL_SERVER_ERROR","message":"Failed to serialize error response."}`)
	}
}

/*
-------------------------------------
    Utilities for Checking Errors
-------------------------------------
*/

// IsAppError checks if err is an *AppError and returns it.
// If not, it wraps the error in a generic *AppError with CodeInternal.
func IsAppError(err error) *AppError {
	if err == nil {
		return nil // Or return a specific "no error" AppError if your system expects it
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	// If it's not an AppError, wrap it as a generic internal error.
	// We need to provide a logger here, which is a design decision.
	// For simplicity, I'm passing nil and assuming LogError will handle it.
	// In a real app, you might have a default logger or pass it from the handler.
	return NewAppError(CodeInternal, "An unexpected internal error occurred.", "system_error_wrapping", err, nil, nil)
}

// HasCode returns true if error or any wrapped error has the given code.
func HasCode(err error, code string) bool {
	for err != nil {
		if ae, ok := err.(*AppError); ok && ae.Code == code {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

// QUERY MAP ERROR
func QueryError(key string) error {
	return fmt.Errorf("%s query not found in the map", key)
}

func FormatErrMsg(context string, err error) string {
	return fmt.Sprintf("%s error: %+v", context, err)
}

// // ValidationError indicates data failed validation before or during store operation.
// type ValidationError struct {
// 	FieldErrors map[string]string // e.g., {"email": "invalid format", "age": "too young"}
// 	Err         error
// }

// func (e *ValidationError) Error() string {
// 	return fmt.Sprintf("validation failed: %v, details: %v", e.Err, e.FieldErrors)
// }

// // PermissionError indicates insufficient permissions for a store operation.
// type PermissionError struct {
// 	Action string
// 	Err    error
// }

// func (e *PermissionError) Error() string {
// 	return fmt.Sprintf("permission denied for action '%s': %v", e.Action, e.Err)
// }
// func (e *PermissionError) Unwrap() error      { return e.Err }
// func (e *PermissionError) IsStoreError() bool { return true }

// DefaultErrorHandler handles errors, logs, and sends JSON responses
// type DefaultErrorHandler struct {
// 	Logger    Logger
// 	Responder *JSONWriter
// }

// func NewErrorHandler(logger Logger, responder *JSONWriter) *DefaultErrorHandler {
// 	return &DefaultErrorHandler{
// 		Logger:    logger,
// 		Responder: responder,
// 	}
// }

// func HandleNoLoggerInCTX(w http.ResponseWriter, r *http.Request, err error) logging.Logger {
// 	if err != nil {
// 		// Create a new AppError for the logger not found in context
// 		NoLoggerErr("ctx_logger", "logger not found in context, using GetLogger()")
// 		logger, err := GetLogger()
// 		if err != nil {
// 			fmt.Println("logger not found in context and cannot be initialized with GetLogger()")
// 			NoLoggerErr("msg", "GetLogger() failed")
// 			return nil
// 		}

// 		logger.Error("Logger was not found in context, using GetLogger(): %+w", err)

//			return logger
//		}
//		return nil
//	}
//

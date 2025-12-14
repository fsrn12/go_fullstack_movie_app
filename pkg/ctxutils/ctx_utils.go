package ctxutils

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"multipass/pkg/common"
	"multipass/pkg/logging"
	"multipass/pkg/response"
)

type contextKey string

const (
	UserContextKey       = contextKey("user")
	UserIDContextKey     = contextKey("user_id")
	loggerContextKey     = contextKey("logger")
	jsonWriterContextKey = contextKey("json_writer")
	traceIDKey           = contextKey("trace_id")
	requestIDKey         = contextKey("request_id")
)

// SetLoggerAndJWInCtx returns a new context with logger and JSON writer stored.
func SetLoggerAndJWInCtx(ctx context.Context, logger logging.Logger, jw response.Writer) context.Context {
	ctx = context.WithValue(ctx, loggerContextKey, logger)
	ctx = context.WithValue(ctx, jsonWriterContextKey, jw)
	return ctx
}

// GetLogAndJWFromCtx returns logger and jsonWriter from  context
func GetLogAndJWFromCtx(ctx context.Context) (logging.Logger, response.Writer, error) {
	var s strings.Builder
	var errFound bool

	logger, ok := ctx.Value(loggerContextKey).(logging.Logger)
	if !ok {
		s.WriteString("Logger not found in context; ")
		errFound = true
	}
	jw, ok := ctx.Value(jsonWriterContextKey).(response.Writer)
	if !ok {
		s.WriteString("JSONWriter not found in context")
		errFound = true
	}
	if errFound {
		return logger, jw, fmt.Errorf("%s", strings.TrimSpace(s.String()))
	}
	return logger, jw, nil
}

func GetLoggerFromCtx(ctx context.Context) logging.Logger {
	logger, ok := ctx.Value(loggerContextKey).(logging.Logger)
	if !ok {
		return nil
	}
	return logger
}

func GetJSONWriterFromCtx(ctx context.Context) response.Writer {
	jsonWriter, ok := ctx.Value(jsonWriterContextKey).(response.Writer)
	if !ok {
		return nil
	}
	return jsonWriter
}

func SetLoggerInCtx(ctx context.Context, logger logging.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func SetJWInCtx(ctx context.Context, jw response.Writer) context.Context {
	return context.WithValue(ctx, jsonWriterContextKey, jw)
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context, log logging.Logger) (int, error) {
	userID, ok := ctx.Value(UserIDContextKey).(*int)
	if !ok || userID == nil {
		return 0, errors.New("userID context not found in context")
	}

	return *userID, nil
}

// func GetUserID(ctx context.Context, log logging.Logger) (int, error) {
// 	userIDValue := ctx.Value(UserIDContextKey)

// 	// If userID is not found, return a sentinel value (e.g., -1)
// 	if userIDValue == nil {
// 		log.Error("Warning: user_id is missing in context", nil)
// 		return -1, fmt.Errorf("warning: user_id is missing in context: %+v", userIDValue)
// 	}

// 	// Try to type assert to int
// 	userID, ok := userIDValue.(int)
// 	if !ok {
// 		log.Error("Warning: user_id is not of expected type int", nil)
// 		return -1, fmt.Errorf("warning: user_id is not of expected type int: %+v", userID)
// 	}

// 	// Optionally, check if userID is valid
// 	if userID <= 0 {
// 		log.Error("Warning: Invalid user_id (<= 0) in context", nil)
// 		return -1, fmt.Errorf("warning: Invalid user_id (<= 0) in context: %+v", userID)
// 	}

// 	// Successfully retrieved the user ID
// 	return userID, nil
// }

// SetUser sets user context struct in the context
func SetUser(ctx context.Context, log logging.Logger, user *common.UserContext) context.Context {
	if user == nil {
		log.Error("Error: Attempted to set nil user in context", nil)
		return ctx
	}
	return context.WithValue(ctx, UserContextKey, user)
}

// Getuser retrieves user from context
func GetUser(ctx context.Context) (*common.UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*common.UserContext)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	fmt.Println("User: ", user)
	return user, nil
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// SetTraceID sets the trace ID in the request's context
func SetTraceID(ctx context.Context, traceID string) context.Context {
	// Create a new context with the trace ID
	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return traceID
	}

	return generateID()
}

// generateID generates a random ID (16 bytes) in hexadecimal format
func generateID() string {
	// Generate a slice of 16 random bytes (128 bits)
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "error-generating-trace-id"
	}
	return fmt.Sprintf("%x", bytes)
}

// Function to add `trace_id` and `request_id` to the request context
func AddIDsToContext(r *http.Request) *http.Request {
	// Check if the `trace_id` is already in the headers or generate one
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = generateID()
	}

	// Generate a new `request_id` for each request
	requestID := generateID()

	// Add the `trace_id` and `request_id` to the request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, requestIDKey, requestID)

	// Update the request to include the new context
	r = r.WithContext(ctx)

	// Optionally, add these IDs to the response headers (for downstream services)
	r.Header.Set("X-Trace-ID", traceID)
	r.Header.Set("X-Request-ID", requestID)

	return r
}

// Function to extract `trace_id` and `request_id` from the request context
func GetIDsFromContext(r *http.Request) (string, string) {
	// Retrieve the `trace_id` and `request_id` from the context
	traceID, ok1 := r.Context().Value(traceIDKey).(string)
	requestID, ok2 := r.Context().Value(requestIDKey).(string)

	// If either is missing, generate a new ID (this shouldn't happen normally)
	if !ok1 || !ok2 {
		traceID = generateID()
		requestID = generateID()
	}

	return traceID, requestID
}

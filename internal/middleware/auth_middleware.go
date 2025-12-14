package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"multipass/internal/auth/tokens"
	"multipass/internal/cache"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/response"

	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	Tokens      *tokens.TokenManager
	RedisClient *redis.Client
	Logger      logging.Logger
	Responder   response.Writer
}

func NewAuthMiddleware(tokens *tokens.TokenManager, redisClient *redis.Client, logger logging.Logger, responder response.Writer) *AuthMiddleware {
	return &AuthMiddleware{
		Tokens:      tokens,
		RedisClient: redisClient,
		Logger:      logger,
		Responder:   responder,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		op := "AuthMiddleware.Authenticate"
		metaData := common.Envelop{
			"op": op,
			"ip": r.RemoteAddr,
		}

		w.Header().Add("Vary", "Authorization")

		// STEP 1: Check if the Authorization header exists
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.handleError(w, r, nil, MissingAuthHeader, metaData)
			return
		}

		// STEP 2: Check if the Authorization header starts with "Bearer "
		// (case-insensitive)
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			metaData["auth_header_prefix"] = authHeader[:min(len(authHeader), 20)]
			m.handleError(w, r, nil, InvalidAuthHeader, metaData)
			return
		}

		// STEP 3: Extract the token from the Authorization header
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			m.handleError(w, r, nil, MissingToken, metaData)
			return
		}

		// STEP 4: Check Cache for Validated JWT Claims
		var cachedClaim *tokens.CustomClaims
		cacheKey := fmt.Sprintf("jwt_claims:%s", token)

		if cache.GetCache(m.RedisClient, cacheKey, cachedClaim) {
			m.handleSetUserCtxAndNext(w, r, next, cachedClaim)
		}

		// STEP 5: Decode the access token
		claims, err := m.Tokens.ValidateJWT(token)
		if err != nil {
			metaData["token_prefix"] = token[:min(len(token), 20)]
			m.handleError(w, r, err, InvalidToken, metaData)
			return
		}

		// STEP 5: Set user context

		metaData["user_id"] = claims.UserID
		m.Logger.Info("Successfully authenticated user in middleware", "meta", metaData)
		m.handleSetUserCtxAndNext(w, r, next, claims)
	})
}

func (m *AuthMiddleware) handleSetUserCtxAndNext(w http.ResponseWriter, r *http.Request, next http.Handler, claims *tokens.CustomClaims) {
	userCtx := &common.UserContext{
		UserID: claims.UserID,
		Name:   claims.Name,
		Email:  claims.Email,
	}

	ctx := ctxutils.SetUser(r.Context(), m.Logger, userCtx)
	m.Logger.Info("user set in context", "user_id", claims.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
}

const (
	MissingAuthHeader = "missing_auth_header"
	InvalidAuthHeader = "malformed_auth_header_no_bearer_prefix"
	MissingToken      = "empty_or_missing_token_after_processing"
	InvalidToken      = "jwt_validation_failure"
)

func (m *AuthMiddleware) handleError(w http.ResponseWriter, r *http.Request, err error, context string, meta common.Envelop) {
	var appErr *apperror.AppError

	switch context {
	case MissingAuthHeader:
		meta["details"] = MissingAuthHeader
		appErr = apperror.ErrMissingAuth(err, m.Logger, meta)
	case InvalidAuthHeader:
		meta["details"] = InvalidAuthHeader
		appErr = apperror.ErrInvalidAuth(nil, m.Logger, meta)
	case MissingToken:
		meta["details"] = MissingToken
		appErr = apperror.ErrInvalidJWT(nil, m.Logger, meta)
	case InvalidToken:
		meta["details"] = InvalidToken
		appErr = apperror.ErrTokenExpired(err, m.Logger, meta)
	default:
		appErr = apperror.NewAppError(err, "authenticatino failed", "AuthMiddleware.Authenticate", err, m.Logger, meta)
	}

	appErr.WriteJSONError(w, r, m.Responder)
}

func (m *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			op := "auth_middleware.Middleware"
			metaData := common.Envelop{
				"op": op,
			}
			w.Header().Add("Vary", "Authorization")

			// Check if the Authorization header exists
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				metaData["details"] = "missing_auth_header"
				appErr := apperror.ErrMissingAuth(nil, m.Logger, metaData)
				appErr.WriteJSONError(w, r, m.Responder)
				return
			}

			// Check if the Authorization header starts with "Bearer " (case-insensitive)
			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				metaData["details"] = "malformed_auth_header_no_bearer_prefix"
				metaData["auth_header_prefix"] = authHeader[:min(len(authHeader), 20)]
				appErr := apperror.ErrInvalidAuth(nil, m.Logger, metaData)
				appErr.WriteJSONError(w, r, m.Responder)
				return
			}

			// Extract the token from the Authorization header
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				metaData["details"] = "empty_token_after_trim"
				appErr := apperror.ErrInvalidJWT(nil, m.Logger, metaData)
				appErr.WriteJSONError(w, r, m.Responder)
				return
			}

			// Decode the access token
			claims, err := m.Tokens.ValidateJWT(token)
			if err != nil {
				metaData["details"] = "jwt_validation_failed"
				metaData["token_prefix"] = token[:min(len(token), 20)]
				appErr := apperror.ErrTokenExpired(err, m.Logger, metaData)
				appErr.WriteJSONError(w, r, m.Responder)
				return
			}

			// Set user context
			userCtx := &common.UserContext{
				UserID: claims.UserID,
				Name:   claims.Name,
				Email:  claims.Email,
			}

			ctx := ctxutils.SetUser(r.Context(), m.Logger, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// if context == MissingAuthHeadher {
// 	meta["details"] = MissingAuthHeadher
// 	appErr = apperror.ErrMissingAuth(err, m.Logger, meta)
// } else if context == InvalidAuthHeader {
// 	meta["details"] = InvalidAuthHeader
// 	appErr = apperror.ErrInvalidAuth(nil, m.Logger, meta)
// } else if context == MissingToken {
// 	meta["details"] = MissingToken
// 	appErr = apperror.ErrInvalidJWT(nil, m.Logger, meta)
// } else if context == InvalidToken {
// 	meta["details"] = InvalidToken
// 	appErr = apperror.ErrTokenExpired(err, m.Logger, meta)
// } else {
// 	appErr = apperror.NewAppError(err, "authenticatino failed", "AuthMiddleware.Authenticate", err, m.Logger, meta)
// }

// func (m *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			op := "auth_middleware.Middleware"
// 			metaData := common.Envelop{
// 				"op": op,
// 			}
// 			w.Header().Add("Vary", "Authorization")

// 			// Check if the Authorization header exists
// 			authHeader := r.Header.Get("Authorization")
// 			if authHeader == "" {
// 				metaData["details"] = "missing_auth_header"
// 				appErr := apperror.ErrMissingAuth(nil, m.Logger, metaData)
// 				appErr.WriteJSONError(w, r, m.Responder)
// 				return
// 			}

// 			// Check if the Authorization header starts with "Bearer " (case-insensitive)
// 			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
// 				metaData["details"] = "malformed_auth_header_no_bearer_prefix"
// 				metaData["auth_header_prefix"] = authHeader[:min(len(authHeader), 20)]
// 				appErr := apperror.ErrInvalidAuth(nil, m.Logger, metaData)
// 				appErr.WriteJSONError(w, r, m.Responder)
// 				return
// 			}

// 			// Extract the token from the Authorization header
// 			token := strings.TrimPrefix(authHeader, "Bearer ")
// 			if token == "" {
// 				metaData["details"] = "empty_token_after_trim"
// 				appErr := apperror.ErrInvalidJWT(nil, m.Logger, metaData)
// 				appErr.WriteJSONError(w, r, m.Responder)
// 				return
// 			}

// 			// Decode the access token
// 			claims, err := m.Tokens.ValidateJWT(token)
// 			if err != nil {
// 				metaData["details"] = "jwt_validation_failed"
// 				metaData["token_prefix"] = token[:min(len(token), 20)]
// 				appErr := apperror.ErrTokenExpired(err, m.Logger, metaData)
// 				appErr.WriteJSONError(w, r, m.Responder)
// 				return
// 			}

// 			// Set user context
// 			userCtx := &common.UserContext{
// 				UserID: claims.UserID,
// 				Name:   claims.Name,
// 				Email:  claims.Email,
// 			}

// 			ctx := ctxutils.SetUser(r.Context(), m.Logger, userCtx)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

// ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
// RATE LIMITING
// func RateLimitMiddleware(redis *redis.Client, limit int, window time.Duration, logger logging.Logger, responder response.Writer) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// We can rate limit by IP or User ID
// 			ip := r.RemoteAddr

// 			// Create a unique key for this IP address and the rate-limit window
// 			key := fmt.Sprintf("rate_limit:%s", ip)

// 			// Check the current number of requests for this IP in the rate-limit window
// 			count, err := redis.Get(context.Background(), key).Int()
// 			// Handle redis.Nil error (key not found)
// 			if err != nil {
// 				if errors.Is(err, redis.NilError) {
// 					// Key doesn't exist (first request), count starts from 0
// 					count = 0
// 				} else {
// 					// If there’s any other error, log and return a server error
// 					logger.Error("Failed to check rate limit from Redis", err)
// 					apperror.NewAppError(http.StatusInternalServerError, "server error", "rate_limit", err, logger, nil).WriteJSONError(w, r, responder)
// 					return
// 				}
// 			}

// 			// If count exceeds the limit, deny access
// 			if count >= limit {
// 				apperror.NewAppError(http.StatusTooManyRequests, "Rate limit exceeded", "rate_limit", err, logger, nil).WriteJSONError(w, r, responder)
// 				return
// 			}
// 			// otherwise, increment count for this key
// 			_, err = redis.Incr(context.Background(), key).Result()
// 			if err != nil {
// 				logger.Error("Failed to increment rate limit counter", err)
// 				apperror.NewAppError(http.StatusInternalServerError, "server error", "rate_limit", err, logger, nil).WriteJSONError(w, r, responder)
// 				return
// 			}

// 			// Set the expiration for the rate-limited key if it doesn't exist
// 			// This ensures that the rate limit window resets after `window` duration
// 			if count == 0 {
// 				_, err := redis.SetEx(context.Background(), key, 0, window).Result()
// 				if err != nil {
// 					logger.Error("Failed to set expiration for rate limit key", err)
// 					apperror.NewAppError(http.StatusInternalServerError, "server error", "rate_limit", err, logger, nil).WriteJSONError(w, r, responder)
// 					return
// 				}

// 			}

// 			// If the rate limit check passes, continue with the request
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
// cookie, err := r.Cookie("access_token")
// if err != nil {
// 	apperror.NewAppError(http.StatusUnauthorized, "missing access token", "authentication_token", err, logger, nil).WriteJSONError(w, r, responder)
// 	return
// }

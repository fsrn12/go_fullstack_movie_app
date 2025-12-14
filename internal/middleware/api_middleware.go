package middleware

import (
	"net/http"
	"time"

	"multipass/pkg/apperror"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/provider"
	"multipass/pkg/response"
)

type ApiMiddleware struct {
	Logger    logging.Logger
	Responder response.Writer
}

func NewApiMiddleware(logger logging.Logger, responder response.Writer) *ApiMiddleware {
	return &ApiMiddleware{
		Logger:    logger,
		Responder: responder,
	}
}

func (a *ApiMiddleware) TraceAndRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add trace and request ID to the request context and headers
		r = ctxutils.AddIDsToContext(r)

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

// logger utils.Logger
func TimingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Getting logger
			logger := ctxutils.GetLoggerFromCtx(r.Context())

			if logger == nil {
				apperror.NoLoggerErr()
				logger, _ = provider.GetLogger()
			}
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info("request completed",
				"method", r.Method, "path", r.URL.Path, "duration_ms", duration.Milliseconds())
		})
	}
}

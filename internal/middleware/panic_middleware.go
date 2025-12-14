package middleware

import (
	"fmt"
	"net/http"

	"multipass/pkg/common"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/response"
)

type PanicRecoverMiddleWare struct {
	Logger    logging.Logger
	Responder response.Writer
}

func NewPanicRecoverMiddleWare(logger logging.Logger, responder response.Writer) *PanicRecoverMiddleWare {
	return &PanicRecoverMiddleWare{
		Logger:    logger,
		Responder: responder,
	}
}

func (m *PanicRecoverMiddleWare) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				m.Logger.Error("Recovered from panic", nil, rec)
				if err := m.Responder.WriteJSON(w, http.StatusInternalServerError, common.Envelop{"error": fmt.Sprintf("panic: %v", rec)}); err != nil {
					m.Logger.Error("Error writing JSON response", err)
				}
			}
		}()

		// Inject logger and JSON writer into context, and update the request with it
		ctx := ctxutils.SetLoggerAndJWInCtx(r.Context(), m.Logger, m.Responder)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

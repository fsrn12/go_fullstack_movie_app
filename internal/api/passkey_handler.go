package api

import (
	"net/http"
	"time"

	"multipass/internal/service"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/cookieutils"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/response"
	"multipass/pkg/utils"
)

const WEBAUTHN_COOKIE_NAME string = "sid"

type WebAuthnHandler struct {
	service service.WebAuthnService
	BaseHandler
}

func NewWebAuthnHandler(service service.WebAuthnService, logger logging.Logger, responder response.Writer) *WebAuthnHandler {
	return &WebAuthnHandler{
		service: service,
		BaseHandler: BaseHandler{
			Logger:       logger,
			Responder:    responder,
			ErrorHandler: apperror.NewBaseErrorHandler(logger, responder),
		},
	}
}

func (h *WebAuthnHandler) WebAuthnRegistrationBeginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "WebAuthnHandler.WebAuthnSignUpStartHandler",
	}

	// STEP 1: GET USER FROM CONTEXT
	ctxUser, err := ctxutils.GetUser(ctx)
	if err != nil {
		metaData["warning"] = "User not found in context"
		if h.ErrorHandler.HandleAppError(w, r, err, "no user in r.Context") {
			return
		}
	}

	// STEP 2: PASS USER.EMAIL to WEBAUTHNSERVICE
	result, err := h.service.WebAuthnSignUpStartService(ctx, ctxUser.Email)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "no user in r.Context") {
			return
		}
	}

	// STEP 3: SET COOKIE
	cookieutils.SetCookie(w, h.Logger, WEBAUTHN_COOKIE_NAME, result.Token, "api/passkey/registerStart", 3600, time.Now().Add(1*time.Hour))

	// STEP 4: RETURN RESPONSE
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": result.Options}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user webauthn start request", "meta", metaData)
}

func (h *WebAuthnHandler) WebAuthRegistrationEndHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "WebAuthnHandler.WebAuthnSignUpStartHandler",
	}

	// STEP 1: GET USER FROM CONTEXT
	ctxUser, err := ctxutils.GetUser(ctx)
	if err != nil {
		metaData["warning"] = "User not found in context"
		if h.ErrorHandler.HandleAppError(w, r, err, "no user in r.Context") {
			return
		}
	}

	// STEP 2: EXTRACT SESSION KEY FROM COOKIE
	cookie, err := cookieutils.GetCookie(r, h.Logger, WEBAUTHN_COOKIE_NAME)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r,
			apperror.ErrUnauthorized(err, h.Logger, metaData), "webauthn_cookie_missing") {
			return
		}
		return
	}

	// STEP 3: WEBAUTHN REGISTRATION END SERVICE
	if err := h.service.WebAuthnSignUpEndService(w, r, cookie.Value, ctxUser.Email); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "no user in r.Context") {
			return
		}
	}

	// STEP 4: RETURN RESPONSE
	resp := struct{ Success bool }{
		Success: true,
	}
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user webauthn end request", "meta", metaData)
}

// LOGIN
func (h *WebAuthnHandler) WebAuthnAuthenticationBeginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "WebAuthnHandler.WebAuthnSignUpStartHandler",
	}

	// STEP 1: DECODE AUTHENTICATION REQUEST
	req, err := utils.DecodeRequest[common.WebAuthnAuthRequest](w, r, "User authentication")
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "from webauthn authentication request") {
			return
		}
	}

	metaData["emial"] = req.Email

	// STEP 2: PASS req.Email to WEBAUTHNSERVICE
	result, err := h.service.WebAuthnLoginStartService(ctx, req.Email)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r, err, "no user in r.Context") {
			return
		}
	}

	// STEP 3: SET COOKIE
	cookieutils.SetCookie(w, h.Logger, WEBAUTHN_COOKIE_NAME, result.Token, "api/passkey/loginStart", 3600, time.Now().Add(1*time.Hour))

	// STEP 4: RETURN RESPONSE
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": result.Options}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user webauthn start request", "meta", metaData)
}

func (h *WebAuthnHandler) WebAuthnAuthenticationEndHandler(w http.ResponseWriter, r *http.Request) {
	metaData := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "WebAuthnHandler.WebAuthnLoginEndHandler",
	}

	// STEP 2: EXTRACT SESSION KEY FROM COOKIE
	cookie, err := cookieutils.GetCookie(r, h.Logger, WEBAUTHN_COOKIE_NAME)
	if err != nil {
		if h.ErrorHandler.HandleAppError(w, r,
			apperror.ErrUnauthorized(err, h.Logger, metaData), "webauthn_cookie_missing") {
			return
		}
		return
	}

	// STEP 3: WEBAUTHN LOGIN END SERVICE
	result, err := h.service.WebAuthnLoginEndService(w, r, cookie.Value)
	if h.ErrorHandler.HandleAppError(w, r, err, "WebAuthnLoginEndService") {
		return
	}

	// STEP 4: SET COOKIE
	cookieutils.SetCookie(w, h.Logger, WEBAUTHN_COOKIE_NAME, result.Token, "/", 3600, time.Now().Add(1*time.Hour))

	// STEP 5: RETURN RESPONSE
	resp := &common.WebAuthnAuthResponse{
		Success: true,
		JWT:     result.JWT,
	}
	if err := h.Responder.WriteJSON(w, http.StatusOK, common.Envelop{"data": resp}); err != nil {
		if h.ErrorHandler.HandleAppError(w, r, apperror.ErrInternalServer(err, h.Logger, metaData), "response writer") {
			return
		}
	}

	h.Logger.Info("successfully processed user webauthn login end request", "meta", metaData)
}

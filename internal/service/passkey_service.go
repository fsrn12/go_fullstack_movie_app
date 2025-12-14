package service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"multipass/internal/auth/tokens"
	"multipass/internal/model"
	"multipass/internal/store"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnService interface {
	WebAuthnSignUpStartService(ctx context.Context, email string) (*common.WebAuthnSignUpResult, error)
	WebAuthnSignUpEndService(w http.ResponseWriter, r *http.Request, token string, email string) error
	WebAuthnLoginStartService(ctx context.Context, email string) (*common.WebAuthnAuthStartResult, error)
	WebAuthnLoginEndService(w http.ResponseWriter, r *http.Request, cookieValue string) (*common.WebAuthnAuthenticationEndResult, error)
}

type PasskeyService struct {
	store        store.PasskeyStore
	webauthn     *webauthn.WebAuthn
	tokenManager *tokens.TokenManager
	logger       logging.Logger
}

func NewPasskeyService(store store.PasskeyStore,
	webauthn *webauthn.WebAuthn,
	tokenManager *tokens.TokenManager,
	logger logging.Logger,
) *PasskeyService {
	return &PasskeyService{
		store:        store,
		webauthn:     webauthn,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

func (s *PasskeyService) WebAuthnSignUpStartService(ctx context.Context, email string) (*common.WebAuthnSignUpResult, error) {
	op := "PasskeyService.WebAuthnSignUpStartService"
	meta := common.Envelop{
		"email": email,
		"op":    op,
	}
	// STEP 1: GET USER FROM DB
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("User not found", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 2: GEGIN WEBAUTHN REGISTRATION
	options, session, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		s.logger.Error("Unable to retrieve email", err)
		return nil, apperror.ErrInternalServer(err, s.logger, meta)
	}

	// STEP 3: GENERATE SESSION KEY & STORE SESSION DATA VALUES
	t, err := s.store.GenSessionID()
	if err != nil {
		s.logger.Error("failed to generate session id: %+w", err)
		return nil, apperror.ErrInternalServer(err, s.logger, meta)
	}

	s.store.SaveSession(t, *session)

	return &common.WebAuthnSignUpResult{
		Options: options,
		Token:   t,
	}, nil
}

func (s *PasskeyService) WebAuthnSignUpEndService(w http.ResponseWriter, r *http.Request, token string, email string) error {
	ctx := r.Context()
	op := "PasskeyService.WebAuthnSignUpEndService"
	meta := common.Envelop{
		"email": email,
		"op":    op,
	}
	// STEP 1: GET SESSION
	session, _ := s.store.GetSession(token)

	// STEP 2: FIND (Passkey) USER BY EMAIL
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("User not found", err, "meta", meta)
		return apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 3: FINISH REGISTRATION
	credential, err := s.webauthn.FinishRegistration(user, session, r)
	if err != nil {
		s.logger.Error("Coudln't finish the WebAuthn Registration", err)
		//! CLEAN UP SID COOKIE
		http.SetCookie(w, &http.Cookie{
			Name:  "sid",
			Value: "",
		})
		return apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 4: STORE CREDENTIAL OBJECT
	user.AddCrediential(credential)
	s.store.SaveUser(ctx, *user)

	// STEP 5: DELETE SESSION DATA (IMPORTANT)
	s.store.DeleteSession(token)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: "",
	})

	return nil
}

func (s *PasskeyService) WebAuthnLoginStartService(ctx context.Context, email string) (*common.WebAuthnAuthStartResult, error) {
	op := "PasskeyService.WebAuthnLoginStartService"
	meta := common.Envelop{
		"email": email,
		"op":    op,
	}
	// STEP 1: GET USER FROM DB
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("User not found", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 2: GEGIN WEBAUTHN REGISTRATION
	options, session, err := s.webauthn.BeginLogin(user)
	if err != nil {
		s.logger.Error("Unable to retrieve email", err)
		return nil, apperror.ErrInternalServer(err, s.logger, meta)
	}

	// STEP 3: GENERATE SESSION KEY & STORE SESSION DATA VALUES
	t, err := s.store.GenSessionID()
	if err != nil {
		s.logger.Error("failed to generate session id: %+w", err)
		return nil, apperror.ErrInternalServer(err, s.logger, meta)
	}

	s.store.SaveSession(t, *session)

	return &common.WebAuthnAuthStartResult{
		Options: options,
		Token:   t,
	}, nil
}

func (s *PasskeyService) WebAuthnLoginEndService(w http.ResponseWriter, r *http.Request, cookieValue string) (*common.WebAuthnAuthenticationEndResult, error) {
	// cookieValue = token
	ctx := r.Context()
	op := "PasskeyService.WebAuthnLoginEndService"
	meta := common.Envelop{
		"op": op,
	}

	// STEP 1: GET SESSION DATA STORED BY WEBAUTHNLOGINSTARTSERVICE
	session, _ := s.store.GetSession(cookieValue)

	// STEP 2: CONVERT USER ID TO INT FROM []byte
	userID, err := strconv.Atoi(string(session.UserID))
	if err != nil {
		s.logger.Error("User not found", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	meta["user_id"] = userID

	// STEP 3: FETCH USER BY ID
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("User not found", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 4: FINISH WEBAUTHN LOGIN
	credential, err := s.webauthn.FinishLogin(user, session, r)
	if err != nil {
		s.logger.Error("Coudln't finish login", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 5: CREDENTIAL AUTHENTICATOR CLOANWARNING
	if credential.Authenticator.CloneWarning {
		s.logger.Error("Failed to finish login", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 6: UPDATE CREDENTIALS IN CASE OF SUCCESSFUL LOGIN
	user.UpdateCredential(credential)
	s.store.SaveUser(ctx, *user)

	// STEP 7: DELETE SESSION DATA
	s.store.DeleteSession(cookieValue)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: "",
	})

	// STEP 8: GENERATE NEW SESSION ID (TOKEN)
	token, err := s.store.GenSessionID()
	if err != nil {
		s.logger.Error("Couldn't generate session", err, "meta", meta)
		return nil, apperror.ErrBadRequest(err, s.logger, meta)
	}

	// STEP 9: SAVE WEBAUTHN SESSION
	s.store.SaveSession(token, webauthn.SessionData{
		Expires: time.Now().Add(time.Hour),
	})

	// STEP 10: GENERATE JWT
	jwt, err := s.tokenManager.CreateJWT(&model.User{
		ID:    userID,
		Email: user.Name,
		Name:  user.DisplayName,
	})
	if err != nil {
		return nil, apperror.ErrInternalServer(err, s.logger, meta)
	}

	// STEP 11: RETURN RESULT BACK TO HANDLER
	return &common.WebAuthnAuthenticationEndResult{
		Token: token,
		JWT:   jwt,
	}, nil
}

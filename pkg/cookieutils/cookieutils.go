package cookieutils

import (
	"errors"
	"net/http"
	"time"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
)

// SetCookie create and sets an httpOnly cookie
func SetCookie(w http.ResponseWriter, log logging.Logger, name, value, path string, maxAge int, ttl time.Time) {
	var expires time.Time
	if maxAge > 0 {
		expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	} else {
		expires = ttl
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	log.Debug("Cookie set", "name", cookie.Name, "maxAge", cookie.MaxAge, "path", cookie.Path, "expires", cookie.Expires, "secure", cookie.Secure, "httponly", cookie.HttpOnly)
}

// GetCookie retrieves cookie by name
func GetCookie(r *http.Request, log logging.Logger, cookieName string) (*http.Cookie, error) {
	meta := common.Envelop{
		"method": r.Method,
		"path":   r.URL.Path,
		"op":     "GetCookie",
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			log.Error("Cookie missing in request: %+w", err)
			return nil, apperror.ErrCookieNotFound(err, log, meta, cookieName)
		}
		log.Error("failed to extract cookie from request: %+w", err)
		return nil, apperror.ErrInternalServer(err, log, meta)
	}

	if cookie.Value == "" {
		log.Warn("invalid or missing value from cookie", "meta", meta)
		return nil, apperror.ErrBadRequest(err, log, meta)
	}

	log.Info("Cookie extracted successfully", "cookie_name", cookie.Name, "meta", meta)
	return cookie, nil
}

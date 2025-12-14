package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"multipass/pkg/logging"
)

type Writer interface {
	WriteJSON(w http.ResponseWriter, status int, data any) error
	// WriteCookie(w http.ResponseWriter, cookie *http.Cookie)
	// WriteAuthCookie(w http.ResponseWriter, name string, value string, path string, maxAge int, ttl time.Time)
}

type JSONWriter struct {
	Logger logging.Logger
}

func NewJSONWriter(logger logging.Logger) Writer {
	if logger == nil {
		// Log a message or panic
		panic("Logger is nil in NewJSONWriter")
	}
	return &JSONWriter{
		Logger: logger,
	}
}

// func (j *JSONWriter) WriteJSON(w http.ResponseWriter, status int, data common.Envelop) error {
func (j *JSONWriter) WriteJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		j.Logger.Error("Failed to marshal JSON", err)
		return fmt.Errorf("failed to marshal data:  %+w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if len(js) == 0 {
		js = []byte("{}")
	}

	js = append(js, '\n')
	_, err = w.Write(js)
	if err != nil {
		// IMPORTANT: Log the error, but do NOT return it from the function.
		// Headers have already been sent, so the caller cannot do anything with this error.
		j.Logger.Errorf("Error writing JSON response to client: %+w", err)
		// Returning 'nil' here indicates that the method completed its attempt to write,
		// even if the underlying network write ultimately failed after headers were sent.
		// return fmt.Errorf("failed writing response: %+w", err)
	}

	// j.Logger.Info("JSON response sent", "status", status)
	return nil
}

// func (j *JSONWriter) WriteAuthCookie(w http.ResponseWriter, name string, value string, path string, maxAge int, ttl time.Time) {
// 	cookie := &http.Cookie{
// 		Name:     name,
// 		Value:    value,
// 		Expires:  ttl,
// 		MaxAge:   maxAge,
// 		Path:     path,
// 		HttpOnly: true,
// 		Secure:   true,
// 		SameSite: http.SameSiteLaxMode,
// 	}
// 	http.SetCookie(w, cookie)
// 	j.Logger.Debug("Cookie set", "name", cookie.Name, "maxAge", cookie.MaxAge, "path", cookie.Path, "expires", cookie.Expires, "secure", cookie.Secure, "httponly", cookie.HttpOnly)
// }

// func (j *JSONWriter) WriteCookie(w http.ResponseWriter, cookie *http.Cookie) {
// 	http.SetCookie(w, cookie)
// 	j.Logger.Debug("Cookie set", "name", cookie.Name, "path", cookie.Path, "expires", cookie.Expires, "secure", cookie.Secure, "httponly", cookie.HttpOnly)
// }

// func (j *JSONWriter) WriteJSON(w http.ResponseWriter, status int, args ...any) error {
// 	data := helpers.MakeEnvelop(args...)
// 	js, err := json.MarshalIndent(data, "", "  ")
// 	if err != nil {
// 		j.Logger.Error("Failed to marshal JSON", err)
// 		return fmt.Errorf("failed to marshal data:  %+w", err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)

// 	if len(js) == 0 {
// 		js = []byte("{}")
// 	}

// 	js = append(js, '\n')
// 	_, err = w.Write(js)
// 	if err != nil {
// 		j.Logger.Errorf("error writing response: %+v\n", err)
// 		return fmt.Errorf("failed writing response: %+v", err)
// 	}

// 	j.Logger.Info("JSON response sent", "status", status)
// 	return nil
// }

// func MakeEnvelop(args ...interface{}) Envelop {
// 	if len(args)%2 != 0 {
// 		panic("metadata arguments must be in key-value pairs (e.g., 'key1', value1, 'key2', value2, ...)")
// 	}

// 	envelop := make(Envelop)
// 	for i := 0; i < len(args); i += 2 {
// 		key, ok := args[i].(string)
// 		if !ok {
// 			panic(fmt.Sprintf("Invalid key type at position %d. Expected string, but got: %T", i, args[i]))
// 		}
// 		envelop[key] = args[i+1]
// 	}
// 	return envelop
// }

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/ctxutils"
	"multipass/pkg/provider"
)

func DecodeRequest[T any](w http.ResponseWriter, r *http.Request, requestType string) (*T, error) {
	const Max_Body_Size int64 = 1 << 20 // MB

	// GET LOGGER & JSONWriter
	logger, jw, err := ctxutils.GetLogAndJWFromCtx(r.Context())
	if err != nil {
		apperror.NoLoggerErr()
		if logger == nil {
			logger, _ = provider.GetLogger()
		}
		if jw == nil {
			jw, _ = provider.GetWriter()
		}
	}

	errMeta := common.Envelop{
		"method":       r.Method,
		"path":         r.URL.Path,
		"request_type": requestType,
	}

	var appErr *apperror.AppError

	// Verify Content-Type
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		errMeta["status"] = http.StatusUnsupportedMediaType
		errMeta["msg"] = fmt.Sprintf("invalid request payload, expected 'application/json', but got %s", r.Header.Get("Content-Type"))

		logger.Error("wrong type of request header",
			nil, errMeta)

		appErr = apperror.ErrInvalidContentType(nil, logger, errMeta)
		appErr.WriteJSONError(w, r, jw)
		return nil, appErr
	}

	// Limit the body size to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, Max_Body_Size)

	// Use the JSON decoder for parsing the request
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var req T
	err = d.Decode(&req)
	if err != nil {
		// Error in decoding, possibly malformed JSON
		errMeta["status"] = http.StatusBadRequest
		logger.Errorf("failed to decode:%+w", err, errMeta)
		appErr = apperror.NewAppError(http.StatusBadRequest, "invalid request payload", "decoding_json", err, logger, errMeta)
		appErr.WriteJSONError(w, r, jw)
		return nil, appErr
	}

	if d.More() {
		// Check if there is extra unexpected data
		errMeta["status"] = http.StatusBadRequest
		logger.Errorf("JSON request contains extra unexpected data:%+w", err, errMeta)
		appErr = apperror.ErrInvalidRequestPayload(nil, logger, errMeta)
		appErr.WriteJSONError(w, r, jw)
		return nil, appErr
	}

	// Successfully decoded the request body
	return &req, nil
}

// var req RegisterRequest
// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// }

// later matching generic response encoder

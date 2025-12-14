package apperror

import (
	"errors"
	"fmt"
	"net/http"

	// Your common Logger and Writer types
	// Your context utility functions
	"multipass/pkg/common"
	"multipass/pkg/ctxutils"
	"multipass/pkg/logging"
	"multipass/pkg/response"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	// For pgx.ErrNoRows
	// For pgconn.PgError (PostgreSQL specific errors)
)

// --- Error Handler Interface and Implementation ---

type ErrorHandler interface {
	HandleAppError(w http.ResponseWriter, r *http.Request, err error, context string) bool
}

type BaseErrorHandler struct {
	Logger    logging.Logger
	Responder response.Writer
}

func NewBaseErrorHandler(logger logging.Logger, responder response.Writer) ErrorHandler {
	return &BaseErrorHandler{
		Logger:    logger,
		Responder: responder,
	}
}

// HandleStoreError is the primary method for handling store-related errors within the HTTP layer.
// It translates specific Go errors (including pgx errors) into common.AppError structure.
func (h *BaseErrorHandler) HandleAppError(w http.ResponseWriter, r *http.Request, err error, context string) bool {
	if err == nil {
		return false // No error to handle
	}

	// 1. Prepare base metadata from the request context
	baseMetadata := common.Envelop{
		"method":  r.Method,
		"path":    r.URL.Path,
		"context": context, // The context from the caller (e.g., "creating user", "fetching movie")
	}

	user, _ := ctxutils.GetUser(r.Context())
	traceID, requestID := ctxutils.GetIDsFromContext(r)

	if user != nil && user.UserID != 0 {
		addIfNotEmpty(baseMetadata, "user_id", user.UserID)
	}
	addIfNotEmpty(baseMetadata, "trace_id", traceID)
	addIfNotEmpty(baseMetadata, "request_id", requestID)

	var finalAppError *AppError

	// 2. Determine the AppError based on the incoming 'err' type
	// This is the core logic: prioritizing specific errors.

	// A. If 'err' is already one of our custom *AppError types, use it directly.
	if appErr, ok := err.(*AppError); ok {
		finalAppError = appErr
	} else if nfe := new(NotFoundError); errors.As(err, &nfe) {
		// B. Handle specific custom error types that might not have been wrapped into *AppError yet (less common if store/service wrap correctly).
		finalAppError = NewAppError(CodeNotFound, ErrNotFoundMsg, context, err, h.Logger,
			common.Envelop{"resource_name": nfe.ResourceName, "resource_id": nfe.ResourceID, "error_detail": nfe.Error()})
	} else if ve := new(ValidationError); errors.As(err, &ve) {
		finalAppError = NewAppError(CodeUnprocessable, ErrValidationMsg, context, err, h.Logger,
			common.Envelop{"field_errors": ve.FieldErrors, "error_detail": ve.Error()})
	} else if ive := new(InvalidInputError); errors.As(err, &ive) {
		finalAppError = NewAppError(CodeBadRequest, fmt.Sprintf("Invalid input for field '%s': %s", ive.Field, ive.Reason), context, err, h.Logger,
			common.Envelop{"field": ive.Field, "reason": ive.Reason})
	} else if pve := new(PolicyViolationError); errors.As(err, &pve) {
		finalAppError = NewAppError(CodeForbidden, fmt.Sprintf("Policy violation: %s", pve.Details), context, err, h.Logger,
			common.Envelop{"policy_name": pve.PolicyName, "details": pve.Details})
	} else if ce := new(ConflictError); errors.As(err, &ce) {
		finalAppError = NewAppError(CodeConflict, ErrConflictMsg, context, err, h.Logger,
			common.Envelop{"reason": ce.Reason, "error_detail": ce.Error()})
	} else if tre := new(TransientError); errors.As(err, &tre) {
		finalAppError = NewAppError(CodeServiceUnavailable, ErrServiceUnavailableMsg, context, err, h.Logger,
			common.Envelop{"error_detail": tre.Error()})
	} else if pe := new(PermissionError); errors.As(err, &pe) {
		finalAppError = NewAppError(CodeForbidden, ErrForbiddenMsg, context, err, h.Logger,
			common.Envelop{"action": pe.Action, "error_detail": pe.Error()})
	} else if rme := new(ResourceModificationError); errors.As(err, &rme) {
		finalAppError = NewAppError(CodeUnprocessable, fmt.Sprintf("Failed to %s resource.", rme.ModificationType), context, err, h.Logger,
			common.Envelop{"resource_name": rme.ResourceName, "resource_id": rme.ResourceID, "modification_type": rme.ModificationType, "error_detail": rme.Error()})
	} else if dbe := new(DBError); errors.As(err, &dbe) {
		// Handle generic DBError from store layer
		finalAppError = NewAppError(CodeDatabaseError, ErrInternalServerMsg, context, err, h.Logger,
			common.Envelop{"operation": dbe.Operation, "sql_hint": dbe.SQL, "error_detail": dbe.Error()})
	} else if errors.Is(err, pgx.ErrNoRows) {
		// C. Handle specific raw pgx/pgconn errors as a fallback if not converted in store layer
		finalAppError = NewAppError(CodeNotFound, ErrNotFoundMsg, context, err, h.Logger,
			common.Envelop{"pgx_error_type": "ErrNoRows"})
	} else if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) {
		h.Logger.Error("Detected pgconn.PgError (unhandled by higher layers)", err,
			"SQLSTATE", pgErr.Code, "Message", pgErr.Message, "Detail", pgErr.Detail,
			"Hint", pgErr.Hint, "ConstraintName", pgErr.ConstraintName, "TableName", pgErr.TableName,
			"ColumnName", pgErr.ColumnName, "base_metadata", baseMetadata,
		)
		switch pgErr.Code {
		case "23505":
			finalAppError = NewAppError(CodeConflict, ErrDuplicateEntryMsg, context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "constraint": pgErr.ConstraintName, "table": pgErr.TableName, "column": pgErr.ColumnName})
		case "23503":
			finalAppError = NewAppError(CodeUnprocessable, "Related resource does not exist or cannot be deleted.", context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "constraint": pgErr.ConstraintName, "table": pgErr.TableName, "column": pgErr.ColumnName})
		case "22001", "22P02":
			finalAppError = NewAppError(CodeUnprocessable, "The provided data format is incorrect or too long.", context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "column": pgErr.ColumnName, "detail": pgErr.Detail})
		case "25P02":
			finalAppError = NewAppError(CodeServiceUnavailable, "Transaction failed. Please try again or contact support.", context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "pg_message": pgErr.Message, "detail": pgErr.Detail})
		case "40P01":
			finalAppError = NewAppError(CodeConflict, "A concurrent operation caused a deadlock. Please try again.", context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "pg_message": pgErr.Message})
		case "08000", "08003", "08006", "08P01":
			finalAppError = NewAppError(CodeServiceUnavailable, ErrServiceUnavailableMsg, context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "pg_message": pgErr.Message})
		case "42P01", "42703":
			finalAppError = NewAppError(CodeInternal, ErrInternalServerMsg, context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "pg_message": pgErr.Message, "detail": pgErr.Detail, "hint": pgErr.Hint, "internal_query": pgErr.InternalQuery})
		default:
			finalAppError = NewAppError(CodeDatabaseError, ErrInternalServerMsg, context, err, h.Logger,
				common.Envelop{"pg_code": pgErr.Code, "pg_message": pgErr.Message, "detail": pgErr.Detail})
		}
	} else {
		// D. Catch-all for any other generic Go `error`. Treat as an internal server error.
		h.Logger.Error("Detected unhandled generic error (not AppError, pgError, or pgx.ErrNoRows)", err, "base_metadata", baseMetadata)
		finalAppError = NewAppError(CodeInternal, ErrInternalServerMsg, context, err, h.Logger, baseMetadata)
	}

	// 3. Finalize the AppError before logging and responding
	// Ensure logger is set
	if finalAppError.Logger == nil {
		finalAppError.Logger = h.Logger
	}

	// Customize the message for generic internal server errors for client consumption
	if finalAppError.Code == CodeInternal && finalAppError.Message == ErrInternalServerMsg {
		finalAppError.Message = fmt.Sprintf("An unexpected error occurred processing your request at %s. Please try again.", context)
	}

	// Merge base metadata into the final AppError's metadata, if keys don't already exist.
	if finalAppError.Metadata == nil {
		finalAppError.Metadata = common.Envelop{}
	}
	for k, v := range baseMetadata {
		if _, exists := finalAppError.Metadata[k]; !exists {
			finalAppError.Metadata[k] = v
		}
	}

	// 4. Log the error internally and write the HTTP JSON response
	finalAppError.LogError()
	finalAppError.WriteJSONError(w, r, h.Responder)
	return true
}

// addIfNotEmpty helper function (your current implementation is robust and good)
func addIfNotEmpty(metadata common.Envelop, key string, value any) {
	if value == nil {
		return
	}
	// Handle different types for zero-value check.
	// Assuming common.User has a meaningful zero value check for UserID if it's a struct.
	// For scalar types, your existing logic is fine.
	switch v := value.(type) {
	case string:
		if v != "" {
			metadata[key] = v
		}
	case int:
		if v != 0 {
			metadata[key] = v
		}
	// ... include other integer types (int32, int64), floats, bool as you had ...
	default:
		// For other types, especially custom structs or interfaces, a direct
		// check for non-nil or specific field values might be needed if they can be "empty"
		// in a non-zero-value way. For simple cases, just add it if not nil.
		metadata[key] = value
	}
}

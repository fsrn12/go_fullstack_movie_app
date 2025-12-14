package apperror

import (
	"fmt"
)

// --- Your Custom Store Error Types (as provided in your context) ---
// These are used by your store layer to wrap database/external errors
// into more specific application-level errors before returning them.
// The HandleStoreError then uses errors.As to identify these.

type StoreError interface {
	error
	IsStoreError() bool // Marker method
}

// NotFoundError indicates a resource was not found in the store.
type NotFoundError struct {
	ResourceName string
	ResourceID   string
	Err          error // Original underlying error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource '%s' with ID '%s' not found: %v", e.ResourceName, e.ResourceID, e.Err)
}
func (e *NotFoundError) Unwrap() error      { return e.Err }
func (e *NotFoundError) IsStoreError() bool { return true }

// ConflictError indicates a conflict, e.g., duplicate entry, optimistic locking failure.
type ConflictError struct {
	Reason string // e.g., "duplicate_key", "optimistic_lock_failed"
	Err    error
}

func NewConflictError(reason string, err error) *ConflictError {
	return &ConflictError{
		Reason: reason,
		Err:    err,
	}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict detected: %s, original error: %v", e.Reason, e.Err)
}
func (e *ConflictError) Unwrap() error      { return e.Err }
func (e *ConflictError) IsStoreError() bool { return true }

// DBError indicates a generic database operation failure (e.g., connection, query execution).
type DBError struct {
	Operation string // e.g., "select", "insert", "transaction"
	SQL       string // The problematic SQL (be careful with sensitive data)
	Err       error  // Original database driver error
	Msg       string
}

func (e *DBError) Error() string {
	return fmt.Sprintf("database error during %s operation: %v", e.Operation, e.Err)
}

func NewDBError(op string, sql string, err error, msg string) *DBError {
	return &DBError{
		Operation: op,
		SQL:       sql,
		Err:       err,
		Msg:       msg,
	}
}
func (e *DBError) Unwrap() error      { return e.Err }
func (e *DBError) IsStoreError() bool { return true }

// TransientError indicates a temporary, retryable error (e.g., deadlock, transient network issue).
type TransientError struct {
	Err error
}

func (e *TransientError) Error() string {
	return fmt.Sprintf("transient error, please retry: %v", e.Err)
}
func (e *TransientError) Unwrap() error      { return e.Err }
func (e *TransientError) IsStoreError() bool { return true }

// ResourceModificationError (kept as per your provided snippet)
type ResourceModificationError struct {
	ResourceName     string
	ResourceID       string
	ModificationType string
	Err              error
}

func (e *ResourceModificationError) Error() string {
	return fmt.Sprintf("error %sing resource, please retry: %v", e.ModificationType, e.Err)
}
func (e *ResourceModificationError) Unwrap() error      { return e.Err }
func (e *ResourceModificationError) IsStoreError() bool { return true } // Assuming it's a store error

// ValidationError (kept as per your provided snippet)
type ValidationError struct {
	FieldErrors map[string]string // e.g., {"email": "invalid format", "age": "too young"}
	Err         error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v, details: %v", e.Err, e.FieldErrors)
}
func (e *ValidationError) Unwrap() error      { return e.Err } // Added Unwrap for consistency
func (e *ValidationError) IsStoreError() bool { return true }  // Assuming it's a store error

// PermissionError (kept as per your provided snippet)
type PermissionError struct {
	Action string
	Err    error
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for action '%s': %v", e.Action, e.Err)
}
func (e *PermissionError) Unwrap() error      { return e.Err }
func (e *PermissionError) IsStoreError() bool { return true }

type InvalidInputError struct {
	Field  string
	Reason string
	Err    error
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input for %s: %s", e.Field, e.Reason)
}
func (e *InvalidInputError) Unwrap() error { return e.Err }

type PolicyViolationError struct {
	PolicyName string
	Details    string
	Err        error
}

func (e *PolicyViolationError) Error() string {
	return fmt.Sprintf("policy violation (%s): %s", e.PolicyName, e.Details)
}
func (e *PolicyViolationError) Unwrap() error { return e.Err }

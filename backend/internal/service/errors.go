package service

import "errors"

// Common service errors shared across services.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict indicates a uniqueness/duplicate constraint violation.
	ErrConflict = errors.New("resource already exists")

	// ErrInvalidState indicates the resource is not in a valid state for the operation.
	ErrInvalidState = errors.New("invalid resource state for this operation")

	// ErrForbidden indicates the user does not have permission.
	ErrForbidden = errors.New("permission denied")

	// ErrExternalService indicates an external service (LLM, WeChat, etc.) call failed.
	ErrExternalService = errors.New("external service error")

	// ErrRateLimited indicates the operation was rate-limited.
	ErrRateLimited = errors.New("rate limited")
)

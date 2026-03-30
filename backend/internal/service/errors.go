package service

import "errors"

// Common service errors shared across services.
var (
	ErrNotFound = errors.New("resource not found")
)

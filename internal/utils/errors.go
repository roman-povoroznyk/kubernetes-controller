package utils

import (
	"fmt"
)

// ErrorType represents different types of errors in the system
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeExternal   ErrorType = "external"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Underlying error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s (underlying: %v)", e.Type, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Underlying
}

// NewValidationError creates a new validation error
func NewValidationError(message string, underlying error) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		Underlying: underlying,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, underlying error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		Underlying: underlying,
	}
}

// NewExternalError creates a new external error (e.g., Kubernetes API errors)
func NewExternalError(message string, underlying error) *AppError {
	return &AppError{
		Type:       ErrorTypeExternal,
		Message:    message,
		Underlying: underlying,
	}
}

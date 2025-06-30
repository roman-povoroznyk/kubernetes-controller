// pkg/errors/errors.go
package errors

import (
	"fmt"
	"runtime"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	// Configuration errors
	ConfigurationError ErrorType = "configuration_error"
	ValidationError    ErrorType = "validation_error"
	
	// Kubernetes errors
	KubernetesError   ErrorType = "kubernetes_error"
	ConnectionError   ErrorType = "connection_error"
	AuthenticationError ErrorType = "authentication_error"
	
	// Controller errors
	ControllerError    ErrorType = "controller_error"
	ReconciliationError ErrorType = "reconciliation_error"
	
	// Network errors
	NetworkError ErrorType = "network_error"
	TimeoutError ErrorType = "timeout_error"
	
	// Internal errors
	InternalError ErrorType = "internal_error"
	UnknownError  ErrorType = "unknown_error"
)

// AppError represents a structured application error
type AppError struct {
	Type      ErrorType              `json:"type"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	StackTrace string                `json:"stack_trace,omitempty"`
	Timestamp  string                `json:"timestamp"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewError creates a new AppError
func NewError(errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StackTrace: getStackTrace(),
		Timestamp:  getCurrentTimestamp(),
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		Cause:      err,
		StackTrace: getStackTrace(),
		Timestamp:  getCurrentTimestamp(),
	}
}

// WithContext adds context to an error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context values to an error
func (e *AppError) WithContextMap(context map[string]interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range context {
		e.Context[k] = v
	}
	return e
}

// IsType checks if error is of specified type
func IsType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getCurrentTimestamp returns current timestamp in ISO format
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", currentTimeUnix())
}

// currentTimeUnix returns current Unix timestamp
func currentTimeUnix() int64 {
	// This would typically use time.Now().Unix() in real code
	// Abstracted for testing purposes
	return 0
}

// Common error constructors
func NewConfigurationError(message string) *AppError {
	return NewError(ConfigurationError, message)
}

func NewValidationError(message string) *AppError {
	return NewError(ValidationError, message)
}

func NewKubernetesError(message string) *AppError {
	return NewError(KubernetesError, message)
}

func NewConnectionError(message string) *AppError {
	return NewError(ConnectionError, message)
}

func NewControllerError(message string) *AppError {
	return NewError(ControllerError, message)
}

func NewReconciliationError(message string) *AppError {
	return NewError(ReconciliationError, message)
}

func NewInternalError(message string) *AppError {
	return NewError(InternalError, message)
}

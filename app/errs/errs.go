package errs

import (
	"errors"
	"net/http"
)

// Code is a stable error code to classify errors across the application.
type Code string

const (
	// EInvalid indicates invalid user input.
	EInvalid Code = "invalid"
	// ENotFound indicates missing resource.
	ENotFound Code = "not_found"
	// EConflict indicates a state conflict.
	EConflict Code = "conflict"
	// EInternal indicates an unexpected internal error.
	EInternal Code = "internal"
)

// AppError is a structured application error that carries a code and message.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Op      string `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap supports errors.Is/As.
func (e *AppError) Unwrap() error { return e.Err }

// Wrap converts any error to *AppError preserving code if already typed.
func Wrap(code Code, msg string, err error) *AppError {
	if ae := From(err); ae != nil && err != nil {
		// keep original nested error for stack/trace; override message/code if provided
		return &AppError{Code: code, Message: msg, Err: err}
	}
	return &AppError{Code: code, Message: msg, Err: err}
}

func Invalid(msg string) *AppError  { return &AppError{Code: EInvalid, Message: msg} }
func NotFound(msg string) *AppError { return &AppError{Code: ENotFound, Message: msg} }
func Conflict(msg string) *AppError { return &AppError{Code: EConflict, Message: msg} }
func Internal(msg string, err ...error) *AppError {
	var e error
	if len(err) > 0 {
		e = err[0]
	}
	return &AppError{Code: EInternal, Message: msg, Err: e}
}

// From tries to extract *AppError from any error.
func From(err error) *AppError {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return &AppError{Code: EInternal, Message: err.Error(), Err: err}
}

// HTTPStatus maps error code to HTTP status.
func HTTPStatus(code Code) int {
	switch code {
	case EInvalid:
		return http.StatusBadRequest
	case ENotFound:
		return http.StatusNotFound
	case EConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

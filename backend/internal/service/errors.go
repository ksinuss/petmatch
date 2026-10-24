package service

import "errors"

// Доменные ошибки. Handlers превращают их в HTTP-статусы (см. handlers.respondError).
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
)

// wrap добавляет человекочитаемое сообщение к доменной ошибке,
// сохраняя возможность errors.Is.
type domainError struct {
	kind error
	msg  string
}

func (e *domainError) Error() string { return e.msg }
func (e *domainError) Unwrap() error { return e.kind }

func invalid(msg string) error      { return &domainError{ErrInvalidInput, msg} }
func notFound(msg string) error     { return &domainError{ErrNotFound, msg} }
func forbidden(msg string) error    { return &domainError{ErrForbidden, msg} }
func conflict(msg string) error     { return &domainError{ErrConflict, msg} }
func unauthorized(msg string) error { return &domainError{ErrUnauthorized, msg} }

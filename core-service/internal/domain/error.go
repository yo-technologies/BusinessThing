package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrAlreadyExists is an error for already existing entity
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrNotFound is an error for not found entity
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidArgument is an error for invalid argument
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrUnauthorized is an error for unauthorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is an error for forbidden
	ErrForbidden = errors.New("forbidden")

	// ErrTooManyRequests is an error for too many requests
	ErrTooManyRequests = errors.New("too many requests")

	// ErrInternal is a generic internal error
	ErrInternal = errors.New("internal error")
)

// DomainError - базовая доменная ошибка
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(message string) error {
	return &DomainError{
		Code:    "NOT_FOUND",
		Message: message,
		Err:     ErrNotFound,
	}
}

// NewInvalidArgumentError создает ошибку "неверный аргумент"
func NewInvalidArgumentError(message string) error {
	return &DomainError{
		Code:    "INVALID_ARGUMENT",
		Message: message,
		Err:     ErrInvalidArgument,
	}
}

// NewUnauthorizedError создает ошибку "не авторизован"
func NewUnauthorizedError(message string) error {
	return &DomainError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewForbiddenError создает ошибку "запрещено"
func NewForbiddenError(message string) error {
	return &DomainError{
		Code:    "FORBIDDEN",
		Message: message,
		Err:     ErrForbidden,
	}
}

// NewInternalError создает внутреннюю ошибку
func NewInternalError(message string, err error) error {
	return &DomainError{
		Code:    "INTERNAL",
		Message: message,
		Err:     err,
	}
}

// NewAlreadyExistsError создает ошибку "уже существует"
func NewAlreadyExistsError(message string) error {
	return &DomainError{
		Code:    "ALREADY_EXISTS",
		Message: message,
		Err:     ErrAlreadyExists,
	}
}

// IsNotFoundError проверяет, является ли ошибка "не найдено"
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNotFound)
}

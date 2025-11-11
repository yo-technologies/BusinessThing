package domain

import (
	"errors"
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

	// ErrAwaitingConfirmation indicates the flow paused waiting for user confirmation
	ErrAwaitingConfirmation = errors.New("awaiting confirmation")
)

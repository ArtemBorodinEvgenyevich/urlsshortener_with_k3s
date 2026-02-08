package domain

import "errors"

var (
	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrInvalidSession  = errors.New("invalid session")

	// User errors
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidUser  = errors.New("invalid user")

	// Provider errors
	ErrInvalidProvider = errors.New("invalid provider")
)

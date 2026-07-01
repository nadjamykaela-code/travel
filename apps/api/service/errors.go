package service

import "errors"

var (
	ErrForbidden   = errors.New("forbidden")
	ErrNotFound    = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

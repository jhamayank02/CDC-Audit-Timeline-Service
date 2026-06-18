package errors

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidID           = errors.New("id must be in valid uuid format")
	ErrInvalidSortBy       = errors.New("sortBy must be asc or desc")
)

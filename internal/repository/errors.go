package repository

import "errors"

var (
	ErrNotFound      = errors.New("record not found")
	ErrDatabaseError = errors.New("database error")
)

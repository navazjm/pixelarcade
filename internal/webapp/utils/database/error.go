package database

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrMockDatabase   = errors.New("mock database error")
)

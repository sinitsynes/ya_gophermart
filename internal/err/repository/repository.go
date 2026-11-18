package repository

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrConflict      = errors.New("conflict")
)

type RepoError struct {
	Err     error
	Message string
}

func (e *RepoError) Error() string {
	return e.Message
}

func (e *RepoError) Unwrap() error {
	return e.Err
}

package models

import (
	"errors"
)

var (
	ErrNotFound = errors.New("models: no record found")

	ErrInvalidCredits = errors.New("models: invalid credits")
)

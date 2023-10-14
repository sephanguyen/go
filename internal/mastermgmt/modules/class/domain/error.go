package domain

import "errors"

var (
	ErrNotFound = errors.New("`class_id` not found")
)

package repositories

import (
	"errors"
)

// ErrUnAffected for unexpected value returned by commandTag
var ErrUnAffected = errors.New("unexpected RowsAffected value")

package domain

import "time"

type Teacher struct {
	ID        string
	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt time.Time
}

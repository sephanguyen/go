package domain

import "time"

type Class struct {
	ClassID   string
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

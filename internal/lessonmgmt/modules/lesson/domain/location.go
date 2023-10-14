package domain

import "time"

type Location struct {
	LocationID string
	Name       string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

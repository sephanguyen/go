package domain

import "time"

type UserAccessPath struct {
	UserID     string
	LocationID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

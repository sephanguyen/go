package domain

import (
	"time"
)

type LmList struct {
	LmListID  string
	LmIDs     []string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewLmListDto() LmList {
	return LmList{}
}

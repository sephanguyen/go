package dto

import "github.com/jackc/pgtype"

// BaseEntity represents default timing column like created_at, updated_at and deleted_at.
type BaseEntity struct {
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

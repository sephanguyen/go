package entities

import (
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
)

// BaseEntity includes default timing column like created_at, updated_at and deleted_at
type BaseEntity struct {
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

// Now sets default value to created_at,  updated_at and deleted_at
func (e *BaseEntity) Now() {
	e.CreatedAt.Set(timeutil.Now())
	e.UpdatedAt = e.CreatedAt
	e.DeletedAt.Set(nil)
}

type LearningMaterialBase struct {
	LearningMaterialID pgtype.Text
	TopicID            pgtype.Text
	Name               pgtype.Text
	Type               pgtype.Text
	DisplayOrder       pgtype.Int2
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
}

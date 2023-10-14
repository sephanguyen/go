package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type CourseAccessPathRepo interface {
}

type CourseRepo interface {
}

type CourseTypeRepo interface {
	GetByIDs(ctx context.Context, db database.Ext, courseTypeIDs []string) ([]*CourseType, error)
}

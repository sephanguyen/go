package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type AcademicYearRepo interface {
	Insert(ctx context.Context, db database.QueryExecer, weeks []*AcademicYear) error
}

type AcademicWeekRepo interface {
}

type AcademicClosedDayRepo interface {
}

type LocationRepo interface {
}

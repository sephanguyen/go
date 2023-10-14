package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
)

type CourseAccessPathRepo interface {
	Upsert(ctx context.Context, db database.Ext, courseap []*domain.CourseAccessPath) error
	Delete(ctx context.Context, db database.QueryExecer, courseIDs []string) error
	GetByCourseIDs(ctx context.Context, db database.Ext, courseIDs []string) ([]*domain.CourseAccessPath, error)
	GetAll(ctx context.Context, db database.QueryExecer) ([]*repo.CourseAccessPath, error)
}

type StudentSubscriptionRepo interface {
	GetLocationActiveStudentSubscriptions(ctx context.Context, db database.Ext, courseIDs []string) (map[string][]string, error)
}

type CourseRepo interface {
	UpdateTeachingMethod(ctx context.Context, db database.Ext, courseList []*domain.Course) error
	Upsert(ctx context.Context, db database.Ext, courseList []*domain.Course) error
	Import(ctx context.Context, db database.Ext, courseList []*domain.Course) error
	GetByIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) ([]*domain.Course, error)
	GetAll(ctx context.Context, db database.QueryExecer) ([]*repo.Course, error)
	LinkSubjects(ctx context.Context, db database.Ext, courses []*domain.Course) error
	GetByPartnerIDs(ctx context.Context, db database.QueryExecer, partnerIDs []string) ([]*domain.Course, error)
}

type CourseTypeRepo interface {
	GetByIDs(ctx context.Context, db database.Ext, courseTypeIDs []string) ([]*domain.CourseType, error)
	Import(ctx context.Context, db database.Ext, courseTypes []*domain.CourseType) error
}

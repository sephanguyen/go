package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	course "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	location "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
)

type ClassRepo interface {
	GetByID(ctx context.Context, db database.QueryExecer, id string) (*Class, error)
	Insert(ctx context.Context, db database.QueryExecer, classes []*Class) error
}

type CourseRepo interface {
	GetByID(ctx context.Context, db database.QueryExecer, courseID string) (*course.Course, error)
}

type LocationRepo interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*location.Location, error)
}

type ClassMemberRepo interface {
	UpsertClassMembers(ctx context.Context, db database.QueryExecer, classMembers []*ClassMember) error
	GetByClassIDAndUserIDs(ctx context.Context, db database.QueryExecer, classID string, userIDs []string) (map[string]*ClassMember, error)
}

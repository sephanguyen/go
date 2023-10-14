package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	course "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	location "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
)

type ClassRepo interface {
	GetByID(ctx context.Context, db database.QueryExecer, id string) (*domain.Class, error)
	GetAll(ctx context.Context, db database.QueryExecer) ([]*domain.ExportingClass, error)
	Insert(ctx context.Context, db database.QueryExecer, classes []*domain.Class) error
	UpdateClassNameByID(ctx context.Context, db database.QueryExecer, id, name string) error
	Delete(ctx context.Context, db database.QueryExecer, id string) error
	UpsertClasses(ctx context.Context, db database.Ext, classes []*domain.Class) error
	RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Class, error)
	FindByCourseIDsAndStudentIDs(ctx context.Context, db database.QueryExecer, cs []*domain.ClassWithCourseStudent) ([]*domain.ClassWithCourseStudent, error)
}
type CourseRepo interface {
	GetByID(ctx context.Context, db database.QueryExecer, courseID string) (*course.Course, error)
}

type LocationRepo interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*location.Location, error)
}

type ClassMemberRepo interface {
	DeleteByUserIDAndClassID(ctx context.Context, db database.QueryExecer, userID, classID string) error
	UpsertClassMembers(ctx context.Context, db database.QueryExecer, classMembers []*domain.ClassMember) error
	UpsertClassMember(ctx context.Context, db database.QueryExecer, classMember *domain.ClassMember) error
	GetByUserAndCourse(ctx context.Context, db database.QueryExecer, userID, courseID string) (map[string]*domain.ClassMember, error)
	GetByClassIDAndUserIDs(ctx context.Context, db database.QueryExecer, classID string, userIDs []string) (map[string]*domain.ClassMember, error)
	FindStudentIDWithCourseIDsByClassIDs(ctx context.Context, db database.QueryExecer, classIds []string) ([]string, error)
}

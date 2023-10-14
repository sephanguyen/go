package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"

	"github.com/jackc/pgtype"
)

type ReserveClassRepo interface {
	InsertOne(ctx context.Context, db database.QueryExecer, reserveClassDomain *domain.ReserveClass) error
	DeleteOldReserveClass(ctx context.Context, db database.QueryExecer, studentPackageID, studentID, courseID string) (pgtype.Text, pgtype.Date, error)
	GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentID string) ([]*domain.ReserveClass, error)
	GetByEffectiveDate(ctx context.Context, db database.QueryExecer, date string) ([]*domain.ReserveClass, error)
	DeleteByEffectiveDate(ctx context.Context, db database.QueryExecer, date string) error
}

type ClassRepo interface {
	GetMapClassByIDs(ctx context.Context, db database.Ext, id []string) (map[string]*repo.Class, error)
}

type CourseRepo interface {
	GetMapCourseByIDs(ctx context.Context, db database.Ext, id []string) (map[string]*repo.Course, error)
}

type StudentPackageClassRepo interface {
	GetManyByStudentPackageIDAndStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, queryString string) ([]*repo.StudentPackageClassDTO, map[string]*repo.StudentPackageClassDTO, error)
	GetStudentPackageClassID(studentPackageID, studentID, courseID string) string
}

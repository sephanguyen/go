package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
)

type CourseLocationScheduleRepo interface {
	UpsertMultiCourseLocationSchedule(ctx context.Context, db database.QueryExecer, arrCourseLocationSchedule []*domain.CourseLocationSchedule) *domain.ImportCourseLocationScheduleError
	ExportCourseLocationSchedule(ctx context.Context, db database.QueryExecer) ([]*domain.CourseLocationSchedule, error)
	GetAcademicWeekValid(ctx context.Context, db database.QueryExecer, locationIds []string, dateValid time.Time) (map[string]bool, error)
	GetByCourseIDAndLocationID(ctx context.Context, db database.QueryExecer, courseID, locationID string) (*domain.CourseLocationSchedule, error)
}

package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	course_location_schedule_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	masterdata_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	student_subscription_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
)

type LessonAllocationRepo interface {
	GetLessonAllocation(ctx context.Context, db database.QueryExecer, filter domain.LessonAllocationFilter) ([]*domain.AllocatedStudent, map[string]uint32, error)
	GetByStudentSubscriptionAndWeek(ctx context.Context, db database.QueryExecer, studentID, courseID string, academicWeekID []string) (map[string][]*domain.LessonAllocationInfo, error)
	CountAssignedSlotPerStudentCourse(ctx context.Context, db database.QueryExecer, studentID, courseID string) (uint32, error)
	CountPurchasedSlotPerStudentSubscription(ctx context.Context, db database.QueryExecer, freq uint8, startTime, endTime time.Time, courseID, locationID, studentID string) (uint32, error)
}

type AcademicWeekRepo interface {
	GetByDateRange(ctx context.Context, db database.Ext, locationID string, academicWeeks []string, startDate time.Time, endDate time.Time) ([]*masterdata_domain.AcademicWeek, error)
}

type AcademicYearRepo interface {
	GetCurrentAcademicYear(ctx context.Context, db database.Ext) (*masterdata_domain.AcademicYear, error)
}

type StudentSubscriptionRepo interface {
	GetByStudentSubscriptionID(ctx context.Context, db database.QueryExecer, studentSubscriptionID string) (*student_subscription_domain.StudentSubscription, error)
}

type StudentSubscriptionAccessPathRepo interface {
	FindLocationsByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, studentSubscriptionIDs []string) (map[string][]string, error)
}

type CourseLocationScheduleRepo interface {
	GetByCourseIDAndLocationID(ctx context.Context, db database.QueryExecer, courseID, locationID string) (*course_location_schedule_domain.CourseLocationSchedule, error)
}

type StudentCourseRepo interface {
	GetByStudentCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID, locationID, studentPackageID string) (*student_subscription_domain.StudentCourse, error)
}

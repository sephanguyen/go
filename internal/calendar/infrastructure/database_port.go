package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_payloads "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type DateInfoPort interface {
	GetDateInfoByDateAndLocationID(ctx context.Context, db database.QueryExecer, date time.Time, locationID string) (*dto.DateInfo, error)
	GetDateInfoByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID string) ([]*dto.DateInfo, error)
	GetDateInfoDetailedByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID, timezone string) ([]*dto.DateInfo, error)
	UpsertDateInfo(ctx context.Context, db database.QueryExecer, params *dto.UpsertDateInfoParams) error
	DuplicateDateInfo(ctx context.Context, db database.QueryExecer, params *dto.DuplicateDateInfoParams) error
	GetAllToExport(ctx context.Context, db database.QueryExecer) ([]byte, error)
}

type DateTypePort interface {
	GetDateTypeByID(ctx context.Context, db database.QueryExecer, id string) (*dto.DateType, error)
	GetDateTypeByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*dto.DateType, error)
}

type LocationPort interface {
	GetLocationByID(ctx context.Context, db database.QueryExecer, id string) (*dto.Location, error)
}

type SchedulerPort interface {
	Create(ctx context.Context, db database.QueryExecer, scheduler *dto.CreateSchedulerParams) (string, error)
	Update(ctx context.Context, db database.QueryExecer, scheduler *dto.UpdateSchedulerParams, updatedFields []string) error
	GetByID(ctx context.Context, db database.QueryExecer, schedulerID string) (*dto.Scheduler, error)
	CreateMany(ctx context.Context, db database.QueryExecer, params []*dto.CreateSchedulerParamWithIdentity) (map[string]string, error)
}

type UserPort interface {
	GetStaffsByLocationIDsAndNameOrEmail(ctx context.Context, db database.QueryExecer, locationIDs, filteredTeacherIDs []string, keyword string, limit int) ([]*dto.User, error)
	GetStaffsByLocationAndWorkingStatus(ctx context.Context, db database.QueryExecer, locationID string, workingStatus []string, useUserBasicInfoTable bool) ([]*dto.User, error)
	GetStudentCurrentGradeByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string, useUserBasicInfoTable bool) (map[string]string, error)
}

type LessonPort interface {
	GetLessonWithNamesByID(ctx context.Context, db database.QueryExecer, lessonID string) (*lesson_domain.Lesson, error)
	GetLessonsByLocationStatusAndDateTimeRange(ctx context.Context, db database.QueryExecer, params *lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs) ([]*lesson_domain.Lesson, error)
}

type LessonMemberPort interface {
	GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]lesson_domain.LessonLearners, error)
}

type LessonTeacherPort interface {
	GetTeachersWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]lesson_domain.LessonTeachers, error)
	GetTeachersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]lesson_domain.LessonTeachers, error)
}

type LessonClassroomPort interface {
	GetLessonClassroomsWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]lesson_domain.LessonClassrooms, error)
}

type LessonGroupPort interface {
	ListMediaByLessonArgs(ctx context.Context, db database.QueryExecer, args *lesson_domain.ListMediaByLessonArgs) (media_domain.Medias, error)
}

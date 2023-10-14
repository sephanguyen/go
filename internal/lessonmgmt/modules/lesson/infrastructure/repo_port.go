package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
)

type LessonRepo interface {
	GetLessonByID(ctx context.Context, db database.QueryExecer, id string) (*domain.Lesson, error)
	GetLessonBySchedulerID(ctx context.Context, db database.QueryExecer, schedulerID string) ([]*domain.Lesson, error)
	InsertLesson(ctx context.Context, db database.QueryExecer, lesson *domain.Lesson) (*domain.Lesson, error)
	UpsertLessons(ctx context.Context, db database.QueryExecer, lesson *domain.RecurringLesson) ([]string, error)
	UpdateLesson(ctx context.Context, db database.QueryExecer, lesson *domain.Lesson) (*domain.Lesson, error)
	GetLessonByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Lesson, error)
	UpdateLessonSchedulingStatus(ctx context.Context, db database.Ext, lesson *domain.Lesson) (*domain.Lesson, error)
	Delete(ctx context.Context, db database.QueryExecer, lessonIDs []string) error
	GetFutureRecurringLessonIDs(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error)
	Retrieve(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonListArg) (ret []*domain.Lesson, total uint32, offsetID string, preTotal uint32, err error)
	UpdateSchedulerID(ctx context.Context, db database.Ext, lessonIDs []string, schedulerIDs string) error
	LockLesson(ctx context.Context, db database.Ext, lessonIds []string) error
	GetLessonsTeachingModelGroupByClassIdWithDuration(ctx context.Context, db database.Ext, query *domain.QueryLesson) ([]*domain.Lesson, error)
	UpdateSchedulingStatus(ctx context.Context, db database.QueryExecer, lessonStatus map[string]domain.LessonSchedulingStatus) error
	GetLessonsOnCalendar(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonListOnCalendarArgs) ([]*domain.Lesson, error)
	GenerateLessonTemplate(ctx context.Context, db database.QueryExecer) ([]byte, error)
	GetLessonWithNamesByID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.Lesson, error)
	RemoveZoomLinkOfLesson(ctx context.Context, db database.QueryExecer, zoomOwnerIds []string) error
	RemoveClassDoLinkOfLesson(ctx context.Context, db database.QueryExecer, classDoOwnerIds []string) error
	GetLessonsByLocationStatusAndDateTimeRange(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs) ([]*domain.Lesson, error)
	GenerateLessonTemplateV2(ctx context.Context, db database.QueryExecer) ([]byte, error)
	GetLessonWithSchedulerInfoByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.Lesson, error)
	RemoveZoomLinkByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error
	RemoveClassDoLinkByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error
	GetFutureLessonsByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string, timezone string) ([]*domain.Lesson, error)
	UpdateLessonsTeachingTime(ctx context.Context, db database.Ext, lesson []*domain.Lesson) error
	GetLessonsWithSchedulerNull(ctx context.Context, db database.QueryExecer, limit int, offset int) ([]*repo.Lesson, error)
	GetLessonsWithInvalidSchedulerID(ctx context.Context, db database.QueryExecer) ([]*repo.Lesson, error)
}

type LessonRoomState interface {
	UpsertCurrentMaterial(ctx context.Context, db database.QueryExecer, material *domain.CurrentMaterial) (*domain.CurrentMaterial, error)
}

type CourseRepo interface {
	UpdateEndDateByCourseIDs(ctx context.Context, db database.Ext, courseIDs []string, endDate time.Time) error
	ExportAllCoursesWithTeachingTimeValue(ctx context.Context, db database.QueryExecer, exportCols []exporter.ExportColumnMap) ([]byte, error)
	CheckCourseIDs(ctx context.Context, db database.QueryExecer, ids []string) error
	RegisterCourseTeachingTime(ctx context.Context, db database.QueryExecer, courses domain.Courses) error
}

type LessonTeacherRepo interface {
	GetTeachersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonTeachers, error)
	GetTeachersWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]domain.LessonTeachers, error)
	UpdateLessonTeacherNames(ctx context.Context, db database.QueryExecer, lessonTeachers []*domain.UpdateLessonTeacherName) error
}

type SearchRepo interface {
	BulkUpsert(ctx context.Context, lessonDocs domain.LessonSearchs) (int, error)
	Search(ctx context.Context, params *domain.ListLessonArgs) (ret []*domain.Lesson, total uint32, offsetID string, err error)
	CreateLessonIndex() error
	DeleteLessonIndex() error
}

type LessonMemberRepo interface {
	ListStudentsByLessonArgs(ctx context.Context, db database.QueryExecer, args *domain.ListStudentsByLessonArgs) ([]*domain.User, error)
	SoftDelete(ctx context.Context, db database.QueryExecer, studentID string, lessonIDs []string) error
	GetLessonIDsByStudentCourseRemovedLocation(ctx context.Context, db database.QueryExecer, courseID, studentID string, locationIDs []string) ([]string, error)
	GetLessonMembersInLessons(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*domain.LessonMember, error)
	GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]domain.LessonLearners, error)
	GetLessonsOutOfStudentCourse(ctx context.Context, db database.QueryExecer, sc *user_domain.StudentSubscription) ([]string, error)
	InsertLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.LessonMember) error
	DeleteLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.LessonMember) error
	DeleteLessonMembersByStartDate(ctx context.Context, db database.QueryExecer, studentID string, classID string, startTime time.Time) ([]string, error)
	UpdateLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.UpdateLessonMemberReport) error
	FindByResourcePath(ctx context.Context, db database.QueryExecer, resourcePath string, limit int, offSet int) (*domain.LessonMembers, error)
	FindByID(ctx context.Context, db database.QueryExecer, lessonID, userID string) (*domain.LessonMember, error)
	UpdateLessonMemberNames(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.UpdateLessonMemberName) error
	UpdateLessonMembersFields(ctx context.Context, db database.QueryExecer, e []*domain.LessonMember, updateFields repo.UpdateLessonMemberFields) error
}

type LessonGroupRepo interface {
	ListMediaByLessonArgs(ctx context.Context, db database.QueryExecer, args *domain.ListMediaByLessonArgs) (media_domain.Medias, error)
}

type StudentRepo interface {
	FindStudentProfilesByIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*usermgmt_entities.LegacyStudent, error)
}

type UserRepo interface {
	Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
}

type DateInfoRepo interface {
	GetDateInfoByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID string) ([]*calendar_dto.DateInfo, error)
}

type LessonReportRepo interface {
	DeleteReportsBelongToLesson(ctx context.Context, db database.Ext, lessonIDs []string) error
	DeleteLessonReportWithoutStudent(ctx context.Context, db database.Ext, lessonID []string) error
}

type ClassroomRepo interface {
	CheckClassroomIDs(ctx context.Context, db database.QueryExecer, classroomIDs []string) error
	ExportAllClassrooms(ctx context.Context, db database.QueryExecer, exportCols []exporter.ExportColumnMap) ([]byte, error)
	UpsertClassrooms(ctx context.Context, db database.QueryExecer, clrs []*domain.Classroom) error
	RetrieveClassroomsByLocationID(ctx context.Context, db database.QueryExecer, params *payloads.GetClassroomListArg) ([]*domain.Classroom, error)
}

type ReallocationRepo interface {
	CancelIfStudentReallocated(ctx context.Context, db database.QueryExecer, studentNewLesson []string) error
	GetReallocatedLesson(ctx context.Context, db database.QueryExecer, lessonMembers []string) ([]*domain.Reallocation, error)
	SoftDelete(ctx context.Context, db database.QueryExecer, studentOriginalLesson []string, isReallocated bool) error
	GetFollowingReallocation(ctx context.Context, db database.QueryExecer, originalLesson string, studentID []string) ([]*domain.Reallocation, error)
	DeleteByOriginalLessonID(ctx context.Context, db database.QueryExecer, originalLesson []string) error
	CancelReallocationByLessonID(ctx context.Context, db database.QueryExecer, newLessonID []string) error
	UpsertReallocation(ctx context.Context, db database.QueryExecer, lessonID string, reallocations []*domain.Reallocation) error
	GetByNewLessonID(ctx context.Context, db database.QueryExecer, studentID []string, newLessonID string) ([]*domain.Reallocation, error)
}

type LessonClassroomRepo interface {
	GetLessonClassroomsWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonClassrooms, error)
	GetOccupiedClassroomByTime(ctx context.Context, db database.QueryExecer, locationIDs []string, lessonID string, starttime, endtime time.Time, timezone string) (*domain.LessonClassrooms, error)
}

package domain

import (
	"context"
	"time"

	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
)

type UserModulePort interface {
	CheckTeacherIDs(ctx context.Context, ids []string) error
	CheckStudentCourseSubscriptions(ctx context.Context, lessonDate time.Time, studentIDWithCourseID ...string) error
}

type MasterDataPort interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*Location, error)
	GetCourseByID(ctx context.Context, db database.Ext, id string) (*Course, error)
	GetClassByID(ctx context.Context, db database.Ext, id string) (*Class, error)
}

type MediaModulePort interface {
	RetrieveMediasByIDs(ctx context.Context, mediaIDs []string) (media_domain.Medias, error)
}

type LessonRepo interface {
	GetLessonByID(ctx context.Context, db database.QueryExecer, id string) (*Lesson, error)
	InsertLesson(ctx context.Context, db database.QueryExecer, lesson *Lesson) (*Lesson, error)
	UpdateLesson(ctx context.Context, db database.QueryExecer, lesson *Lesson) (*Lesson, error)
	UpdateLessonSchedulingStatus(ctx context.Context, db database.Ext, lesson *Lesson) (*Lesson, error)
}

type DateInfoRepo interface {
	GetDateInfoByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID string) ([]*calendar_dto.DateInfo, error)
}

type UserRepo interface {
	GetUserByUserID(ctx context.Context, db database.QueryExecer, userID string) (*user_domain.User, error)
}

type ClassroomRepo interface {
	CheckClassroomIDs(ctx context.Context, db database.QueryExecer, classroomIDs []string) error
}

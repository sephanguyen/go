package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type UserModulePort interface {
	CheckTeacherIDs(ctx context.Context, ids []string) error
	CheckStudentCourseSubscriptions(ctx context.Context, lessonDate time.Time, studentIDWithCourseID ...string) error
	GetUserGroup(ctx context.Context, userID string) (string, error)
}

type UserAccessPathPort interface {
	GetLocationAssignedByUserID(ctx context.Context, db database.QueryExecer, userID []string) (map[string][]string, error)
	Create(ctx context.Context, db database.QueryExecer, userAccessPaths []*domain.UserAccessPath) error
}

type StudentEnrollmentStatusHistoryPort interface {
	Create(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error
}

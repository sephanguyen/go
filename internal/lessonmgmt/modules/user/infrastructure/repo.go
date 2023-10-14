package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
)

type TeacherRepo interface {
	ListByIDs(ctx context.Context, db database.QueryExecer, ids []string) (domain.Teachers, error)
	ListByGrantedLocation(ctx context.Context, db database.QueryExecer) (map[string][]string, error)
}

type StudentSubscriptionRepo interface {
	// GetStudentCourseSubscriptions will get student course subscription using
	// param studentIDWithCourseID, e.g: GetStudentCourseSubscriptions(ctx, db,
	// "student_id_1", "course_id_1", "student_id_1", "course_id_1")
	GetStudentCourseSubscriptions(ctx context.Context, db database.QueryExecer, locationID []string, studentIDWithCourseID ...string) (domain.StudentSubscriptions, error)
	RetrieveStudentSubscription(ctx context.Context, db database.QueryExecer, args *payloads.ListStudentSubScriptionsArgs) ([]*domain.StudentSubscription, uint32, string, uint32, error)
	GetStudentSubscriptionIDByUniqueIDs(ctx context.Context, db database.QueryExecer, subscriptionID, studentID, courseID string) (string, error)
	BulkUpsertStudentSubscription(ctx context.Context, db database.QueryExecer, subList domain.StudentSubscriptions) error
	UpdateMultiStudentNameByStudents(ctx context.Context, db database.QueryExecer, users domain.Users) error
	RetrieveStudentPendingReallocate(ctx context.Context, db database.QueryExecer, params domain.RetrieveStudentPendingReallocateDto) ([]*domain.ReallocateStudent, uint32, error)
	GetStudentCoursesAndClasses(ctx context.Context, db database.QueryExecer, studentID string) (*domain.StudentCoursesAndClasses, error)
	GetAll(ctx context.Context, db database.QueryExecer) ([]*domain.EnrolledStudent, error)
	GetByStudentSubscriptionID(ctx context.Context, db database.QueryExecer, studentSubscriptionID string) (*domain.StudentSubscription, error)
	GetByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, studentSubscriptionID []string) ([]*domain.StudentSubscription, error)
}

type StudentSubscriptionAccessPathRepo interface {
	FindLocationsByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, studentSubscriptionIDs []string) (mapLocationIDBySubscriptionID map[string][]string, err error)
	FindStudentSubscriptionIDsByLocationIDs(ctx context.Context, db database.QueryExecer, locationIds []string) ([]string, error)
	BulkUpsertStudentSubscriptionAccessPath(ctx context.Context, db database.QueryExecer, subList domain.StudentSubscriptionAccessPaths) error
	DeleteByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, subIDList []string) error
}

type ClassRepo interface {
	FindByCourseIDsAndStudentIDs(ctx context.Context, db database.QueryExecer, cs []*class_domain.ClassWithCourseStudent) ([]*class_domain.ClassWithCourseStudent, error)
}

type UserRepo interface {
	GetStudentsManyReferenceByNameOrEmail(ctx context.Context, db database.QueryExecer, keyword string, limit, offset uint32) (domain.Students, error)
	GetUserGroupByUserID(ctx context.Context, db database.QueryExecer, id string) (string, error)
	GetStudentCurrentGradeByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]string, error)
	Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*repo.User, error)
	GetUserByUserID(ctx context.Context, db database.QueryExecer, userID string) (*domain.User, error)
}

type UserBasicInfoRepo interface {
	GetTeachersSameGrantedLocation(ctx context.Context, db database.QueryExecer, query domain.UserBasicInfoQuery) (domain.UsersBasicInfo, error)
	GetUser(ctx context.Context, db database.QueryExecer, userIDs []string) ([]*repo.UserBasicInfo, error)
}

package staff

import (
	"context"

	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	ums "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	ugs "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service/user_group"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/jackc/pgtype"
)

type StaffService struct {
	DB                 database.Ext
	FirebaseAuthClient internal_auth_tenant.TenantClient
	TenantManager      internal_auth_tenant.TenantManager
	FirebaseClient     firebase.AuthClient
	FirebaseUtils      firebase.AuthUtils
	FatimaClient       fpb.SubscriptionModifierServiceClient
	JSM                nats.JetStreamManagement
	UnleashClient      unleashclient.ClientInstance
	Env                string

	DomainUser *ums.DomainUser

	UserModifierService *ums.UserModifierService
	UserGroupV2Service  *ugs.UserGroupService

	SchoolAdminRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, schoolAdmin *entity.SchoolAdmin) error
		SoftDelete(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) error
	}
	TeacherRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error
		SoftDelete(context.Context, database.QueryExecer, pgtype.Text) error
	}
	StaffRepo interface {
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*entity.Staff, error)
		Update(ctx context.Context, db database.QueryExecer, staff *entity.Staff) (*entity.Staff, error)
		Create(ctx context.Context, db database.QueryExecer, staff *entity.Staff) error
	}
	UserRepo interface {
		GetUserRoles(ctx context.Context, db database.QueryExecer, userID string) (entity.DomainRoles, error)
		GetByExternalUserIDs(ctx context.Context, db database.QueryExecer, externalUserIDs []string) (entity.Users, error)
		GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
		GetByEmailsInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
		GetByUserNames(ctx context.Context, db database.QueryExecer, usernames []string) (entity.Users, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error)
	}
	UserGroupRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entity.UserGroup) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, userID pgtype.Text, status pgtype.Text) error
		UpdateOrigin(ctx context.Context, db database.QueryExecer, userID pgtype.Text, isOrigin pgtype.Bool) error
	}
	UserPhoneNumberRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userPhoneNumbers []*entity.UserPhoneNumber) error
	}
	UserAccessPathRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error
		FindLocationIDsFromUserID(ctx context.Context, db database.QueryExecer, userID string) ([]string, error)
	}

	RoleRepo interface {
		GetByUserGroupIDs(ctx context.Context, db database.QueryExecer, userGroupIDs []string) (entity.DomainRoles, error)
	}
}

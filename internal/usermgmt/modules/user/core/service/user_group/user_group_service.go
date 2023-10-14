package user_group

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	ums "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

type UserGroupService struct {
	UserModifierService *ums.UserModifierService
	DB                  database.Ext
	UnleashClient       unleashclient.ClientInstance
	Env                 string
	JSM                 nats.JetStreamManagement

	RoleRepo interface {
		GetRolesByRoleIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*entity.Role, error)
	}

	UserGroupRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entity.UserGroup) error
		DeactivateMultiple(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, groupID pgtype.Text) error
	}

	UserGroupV2Repo interface {
		Create(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroupV2) error
		FindByIDs(ctx context.Context, db database.QueryExecer, userGroupIDs []string) ([]*entity.UserGroupV2, error)
		Find(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) (*entity.UserGroupV2, error)
		Update(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroupV2) error
		FindUserGroupAndRoleByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (map[string][]*entity.Role, error)
	}

	GrantedRoleRepo interface {
		Create(ctx context.Context, db database.QueryExecer, grantedRole *entity.GrantedRole) error
		LinkGrantedRoleToAccessPath(ctx context.Context, db database.QueryExecer, grantedRole *entity.GrantedRole, locationIDs []string) error
		GetByUserGroup(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]*entity.GrantedRole, error)
		Upsert(ctx context.Context, db database.QueryExecer, grantedRoles []*entity.GrantedRole) error
		SoftDelete(ctx context.Context, db database.QueryExecer, grantedRoleIDs pgtype.TextArray) error
	}

	GrantedRoleAccessPathRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, grantedRoleAccessPaths []*entity.GrantedRoleAccessPath) error
	}

	UserGroupsMemberRepo interface {
		GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupMember, error)
		UpsertBatch(ctx context.Context, db database.QueryExecer, userGroupMembers []*entity.UserGroupMember) error
		SoftDelete(ctx context.Context, db database.QueryExecer, userGroupsMembers []*entity.UserGroupMember) error
	}

	LocationRepo interface {
		GetLocationOrg(ctx context.Context, db database.Ext, resourcePath string) (*domain.Location, error)
	}

	UserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error)
		GetUsersByUserGroupID(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]*entity.LegacyUser, error)
		UpdateManyUserGroup(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, userGroup pgtype.Text) error
	}
}

var PlatformPermissionMapper = map[cpb.Platform]map[string]struct{}{
	cpb.Platform_PLATFORM_BACKOFFICE: {
		constant.RoleSchoolAdmin:   struct{}{},
		constant.RoleTeacher:       struct{}{},
		constant.RoleCentreLead:    struct{}{},
		constant.RoleCentreManager: struct{}{},
		constant.RoleCentreStaff:   struct{}{},
		constant.RoleHQStaff:       struct{}{},
		constant.RoleTeacherLead:   struct{}{},
	},
	cpb.Platform_PLATFORM_TEACHER: {
		constant.RoleTeacher: struct{}{},
	},

	cpb.Platform_PLATFORM_LEARNER: {
		constant.RoleStudent: struct{}{},
		constant.RoleParent:  struct{}{},
	},
}

var PlatformPermissionMapperV2 = map[cpb.Platform]map[string]struct{}{
	cpb.Platform_PLATFORM_BACKOFFICE: {
		constant.RoleSchoolAdmin:   struct{}{},
		constant.RoleTeacher:       struct{}{},
		constant.RoleCentreLead:    struct{}{},
		constant.RoleCentreManager: struct{}{},
		constant.RoleCentreStaff:   struct{}{},
		constant.RoleHQStaff:       struct{}{},
		constant.RoleTeacherLead:   struct{}{},
	},
	cpb.Platform_PLATFORM_TEACHER: {
		constant.RoleTeacher:       struct{}{},
		constant.RoleSchoolAdmin:   struct{}{},
		constant.RoleCentreLead:    struct{}{},
		constant.RoleCentreManager: struct{}{},
		constant.RoleCentreStaff:   struct{}{},
		constant.RoleHQStaff:       struct{}{},
		constant.RoleTeacherLead:   struct{}{},
	},

	cpb.Platform_PLATFORM_LEARNER: {
		constant.RoleStudent: struct{}{},
		constant.RoleParent:  struct{}{},
	},
}

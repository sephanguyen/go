package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainUserGroupRepo struct{}

type UserGroupAttribute struct {
	ID             field.String
	Name           field.String
	OrgLocationID  field.String
	IsSystem       field.Boolean
	OrganizationID field.String
}

type UserGroup struct {
	UserGroupAttribute

	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func NewUserGroup(ug entity.DomainUserGroup) *UserGroup {
	now := field.NewTime(time.Now())
	return &UserGroup{
		UserGroupAttribute: UserGroupAttribute{
			ID:             ug.UserGroupID(),
			Name:           ug.Name(),
			OrgLocationID:  ug.OrgLocationID(),
			IsSystem:       ug.IsSystem(),
			OrganizationID: ug.OrganizationID(),
		},
		UpdatedAt: now,
		CreatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (ug *UserGroup) UserGroupID() field.String {
	return ug.UserGroupAttribute.ID
}
func (ug *UserGroup) Name() field.String {
	return ug.UserGroupAttribute.Name
}
func (ug *UserGroup) OrgLocationID() field.String {
	return ug.UserGroupAttribute.OrgLocationID
}
func (ug *UserGroup) IsSystem() field.Boolean {
	return ug.UserGroupAttribute.IsSystem
}
func (ug *UserGroup) OrganizationID() field.String {
	return ug.UserGroupAttribute.OrganizationID
}

func (*UserGroup) TableName() string {
	return "user_group"
}

func (ug *UserGroup) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_group_id",
			"user_group_name",
			"org_location_id",
			"is_system",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&ug.UserGroupAttribute.ID,
			&ug.UserGroupAttribute.Name,
			&ug.UserGroupAttribute.OrgLocationID,
			&ug.UserGroupAttribute.IsSystem,
			&ug.CreatedAt,
			&ug.UpdatedAt,
			&ug.UserGroupAttribute.OrganizationID,
		}
}

func (r *DomainUserGroupRepo) FindUserGroupByRoleName(ctx context.Context, db database.QueryExecer, roleName string) (entity.DomainUserGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserGroupRepo.FindUserGroupByRoleName")
	defer span.End()

	userGroup := &UserGroup{}
	grantedRole := &GrantedRole{}
	fields := database.GetFieldNames(userGroup)
	userGroupFields := make([]string, len(fields))
	for index := range fields {
		userGroupFields[index] = fmt.Sprintf("ug.%s", fields[index])
	}
	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			%s ug
		INNER JOIN %s gr
			ON ug.user_group_id = gr.user_group_id
		INNER JOIN role r
			ON gr.role_id = r.role_id
		WHERE 
			ug.is_system = true AND r.role_name = $1 AND r.deleted_at IS NULL;
	`, strings.Join(userGroupFields, ", "), userGroup.TableName(), grantedRole.TableName())
	if err := db.
		QueryRow(ctx, query, roleName).
		Scan(database.GetScanFields(userGroup, fields)...); err != nil {
		return nil, InternalError{
			RawError: fmt.Errorf("row.Scan: %w", err),
		}
	}

	return userGroup, nil
}

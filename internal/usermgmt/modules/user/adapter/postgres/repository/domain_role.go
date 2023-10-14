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

type DomainRoleRepo struct{}

type RoleAttribute struct {
	RoleID         field.String
	RoleName       field.String
	IsSystem       field.Boolean
	OrganizationID field.String
}

type Role struct {
	RoleAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewRole(role entity.DomainRole) *Role {
	now := field.NewTime(time.Now())
	return &Role{
		RoleAttribute: RoleAttribute{
			RoleID:         role.RoleID(),
			RoleName:       role.RoleName(),
			IsSystem:       role.IsSystem(),
			OrganizationID: role.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func NewNullRole() *Role {
	return NewRole(entity.NullDomainRole{})
}

func (r *Role) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"role_id",
			"role_name",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"is_system"},
		[]interface{}{
			&r.RoleAttribute.RoleID,
			&r.RoleAttribute.RoleName,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
			&r.RoleAttribute.OrganizationID,
			&r.RoleAttribute.IsSystem}
}

func (r *Role) TableName() string {
	return "role"
}
func (r *Role) RoleID() field.String {
	return r.RoleAttribute.RoleID
}
func (r *Role) RoleName() field.String {
	return r.RoleAttribute.RoleName
}
func (r *Role) IsSystem() field.Boolean {
	return r.RoleAttribute.IsSystem
}
func (r *Role) OrganizationID() field.String {
	return r.RoleAttribute.OrganizationID
}

func (r *DomainRoleRepo) GetByUserGroupIDs(ctx context.Context, db database.QueryExecer, userGroupIDs []string) (entity.DomainRoles, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainRoleRepo.GetByUserGroupIDs")
	defer span.End()

	stmt := `
		SELECT
	    r.%s
	  FROM
	    role r
	  INNER JOIN granted_role gr ON r.role_id = gr.role_id
	  WHERE gr.user_group_id = any($1)
			AND gr.deleted_at IS NULL
			AND r.is_system = false`
	role := NewRole(entity.NullDomainRole{})

	fieldNames, _ := role.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ", r."),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(userGroupIDs),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := entity.DomainRoles{}
	for rows.Next() {
		item := NewRole(entity.NullDomainRole{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}

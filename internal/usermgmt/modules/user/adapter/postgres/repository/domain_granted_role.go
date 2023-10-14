package repository

import (
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainGrantedRoleRepo struct{}

type GrantedRoleAttribute struct {
	ID             field.String
	UserGroupID    field.String
	RoleID         field.String
	OrganizationID field.String
}

type GrantedRole struct {
	GrantedRoleAttribute

	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func NewGrantedRole(gl entity.DomainGrantedRole) *GrantedRole {
	now := field.NewTime(time.Now())
	return &GrantedRole{
		GrantedRoleAttribute: GrantedRoleAttribute{
			ID:             gl.ID(),
			UserGroupID:    gl.UserGroupID(),
			RoleID:         gl.RoleID(),
			OrganizationID: gl.OrganizationID(),
		},
		UpdatedAt: now,
		CreatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (gl *GrantedRole) ID() field.String {
	return gl.GrantedRoleAttribute.ID
}
func (gl *GrantedRole) UserGroupID() field.String {
	return gl.GrantedRoleAttribute.UserGroupID
}
func (gl *GrantedRole) OrganizationID() field.String {
	return gl.GrantedRoleAttribute.OrganizationID
}

func (*GrantedRole) TableName() string {
	return "granted_role"
}

func (gl *GrantedRole) FieldMap() ([]string, []interface{}) {
	return []string{
			"granted_role_id",
			"user_group_id",
			"role_id",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&gl.GrantedRoleAttribute.ID,
			&gl.GrantedRoleAttribute.UserGroupID,
			&gl.GrantedRoleAttribute.RoleID,
			&gl.CreatedAt,
			&gl.UpdatedAt,
			&gl.GrantedRoleAttribute.OrganizationID,
		}
}

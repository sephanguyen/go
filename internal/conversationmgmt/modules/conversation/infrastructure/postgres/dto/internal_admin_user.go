package dto

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"

	"github.com/jackc/pgtype"
)

type InternalAdminUserPgDTO struct {
	UserID       pgtype.Text
	VendorUserID pgtype.Text
	IsSystem     pgtype.Bool
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

func (*InternalAdminUserPgDTO) TableName() string {
	return "internal_admin_user"
}

func (dto *InternalAdminUserPgDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "vendor_user_id", "is_system", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&dto.UserID, &dto.VendorUserID, &dto.IsSystem, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt}
	return
}

func (dto *InternalAdminUserPgDTO) ToInternalAdminUserDomain() *domain.InternalAdminUser {
	vendorUser := &domain.InternalAdminUser{
		UserID:       dto.UserID.String,
		VendorUserID: dto.VendorUserID.String,
		IsSystem:     dto.IsSystem.Bool,
	}

	return vendorUser
}

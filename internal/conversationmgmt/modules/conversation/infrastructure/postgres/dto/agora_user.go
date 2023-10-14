package dto

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AgoraUserPgDTO struct {
	UserID      pgtype.Text
	AgoraUserID pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (*AgoraUserPgDTO) TableName() string {
	return "agora_user"
}

func (dto AgoraUserPgDTO) ToChatVendorUserDomain() *domain.ChatVendorUser {
	vendorUser := &domain.ChatVendorUser{
		UserID:       dto.UserID.String,
		VendorUserID: dto.AgoraUserID.String,
	}

	return vendorUser
}

func (dto *AgoraUserPgDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "agora_user_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&dto.UserID, &dto.AgoraUserID, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt}
	return
}

type AgoraUserPgDTOs []*AgoraUserPgDTO

func (dtoArr *AgoraUserPgDTOs) Add() database.Entity {
	e := &AgoraUserPgDTO{}
	*dtoArr = append(*dtoArr, e)

	return e
}

func (dtoArr AgoraUserPgDTOs) ToChatVendorUsersDomain() []*domain.ChatVendorUser {
	chatVendorUsers := make([]*domain.ChatVendorUser, 0)

	for _, item := range dtoArr {
		chatVendorUsers = append(chatVendorUsers, item.ToChatVendorUserDomain())
	}

	return chatVendorUsers
}

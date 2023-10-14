package mappers

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
)

func AgoraUserToCreateUserReq(agoraUser *models.AgoraUser) *chatvendor_dto.CreateUserRequest {
	req := &chatvendor_dto.CreateUserRequest{
		UserID:       agoraUser.UserID.String,
		VendorUserID: agoraUser.AgoraUserID.String,
	}

	return req
}

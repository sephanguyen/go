package mappers

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"

	"github.com/stretchr/testify/assert"
)

func Test_AgoraUserToCreateUserReq(t *testing.T) {
	t.Parallel()

	var agoraUser = &models.AgoraUser{}
	assert.NoError(t, faker.FakeData(agoraUser))

	agoraUserFailure := AgoraUserToCreateUserReq(agoraUser)
	assert.Equal(t, agoraUser.UserID.String, agoraUserFailure.UserID)
	assert.Equal(t, agoraUser.AgoraUserID.String, agoraUserFailure.VendorUserID)
}

package mappers

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
)

func Test_CreateStudentEvtToAgoraUserEnt(t *testing.T) {
	t.Parallel()

	var evt = &upb.EvtUser_CreateStudent{}
	assert.NoError(t, faker.FakeData(evt))

	agoraUser, err := CreateStudentEvtToAgoraUserEnt(evt)
	assert.Nil(t, err)
	assert.Equal(t, agoraUser.UserID.String, evt.StudentId)
	assert.Equal(t, agoraUser.AgoraUserID.String, utils.GetAgoraUserID(evt.StudentId))
}

func Test_CreateParentEvtToAgoraUserEnt(t *testing.T) {
	t.Parallel()

	var evt = &upb.EvtUser_CreateParent{}
	assert.NoError(t, faker.FakeData(evt))

	agoraUser, err := CreateParentEvtToAgoraUserEnt(evt)
	assert.Nil(t, err)
	assert.Equal(t, agoraUser.UserID.String, evt.ParentId)
	assert.Equal(t, agoraUser.AgoraUserID.String, utils.GetAgoraUserID(evt.ParentId))
}

func Test_UpsertStaffEvtToAgoraUserEnt(t *testing.T) {
	t.Parallel()

	var evt = &upb.EvtUpsertStaff{}
	assert.NoError(t, faker.FakeData(evt))

	agoraUser, err := UpsertStaffEvtToAgoraUserEnt(evt)
	assert.Nil(t, err)
	assert.Equal(t, agoraUser.UserID.String, evt.StaffId)
	assert.Equal(t, agoraUser.AgoraUserID.String, utils.GetAgoraUserID(evt.StaffId))
}

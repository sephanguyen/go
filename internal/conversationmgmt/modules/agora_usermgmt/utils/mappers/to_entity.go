package mappers

import (
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
)

func CreateStudentEvtToAgoraUserEnt(createStudentEvt *upb.EvtUser_CreateStudent) (*models.AgoraUser, error) {
	agoraUser := &models.AgoraUser{}
	database.AllNullEntity(agoraUser)

	studentID := createStudentEvt.StudentId
	err := multierr.Combine(
		agoraUser.UserID.Set(studentID),
		agoraUser.AgoraUserID.Set(utils.GetAgoraUserID(studentID)),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: [%v]", err)
	}

	return agoraUser, nil
}

func CreateParentEvtToAgoraUserEnt(createParentEvt *upb.EvtUser_CreateParent) (*models.AgoraUser, error) {
	agoraUser := &models.AgoraUser{}
	database.AllNullEntity(agoraUser)

	parentID := createParentEvt.ParentId
	err := multierr.Combine(
		agoraUser.UserID.Set(parentID),
		agoraUser.AgoraUserID.Set(utils.GetAgoraUserID(parentID)),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: [%v]", err)
	}

	return agoraUser, nil
}

func UpsertStaffEvtToAgoraUserEnt(upsertStaffEvt *upb.EvtUpsertStaff) (*models.AgoraUser, error) {
	agoraUser := &models.AgoraUser{}
	database.AllNullEntity(agoraUser)

	staffID := upsertStaffEvt.StaffId
	err := multierr.Combine(
		agoraUser.UserID.Set(staffID),
		agoraUser.AgoraUserID.Set(utils.GetAgoraUserID(staffID)),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: [%v]", err)
	}

	return agoraUser, nil
}

func AgoraUserEntToAgoraUserFailureEnt(agoraUser *models.AgoraUser) (*models.AgoraUserFailure, error) {
	agoraUserFailure := &models.AgoraUserFailure{}
	database.AllNullEntity(agoraUser)

	err := multierr.Combine(
		agoraUserFailure.UserID.Set(agoraUser.UserID.String),
		agoraUserFailure.AgoraUserID.Set(agoraUser.AgoraUserID.String),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: [%v]", err)
	}

	return agoraUserFailure, nil
}

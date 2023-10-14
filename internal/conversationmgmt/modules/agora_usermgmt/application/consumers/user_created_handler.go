package consumers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/constants"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/infrastructure"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils/mappers"
	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Handle for case:
// + Create student
// + Create parent
type UserCreatedHandler struct {
	DB               database.Ext
	ChatVendorClient chatvendor.ChatVendorClient
	Logger           zap.Logger

	AgoraUserRepo     infrastructure.AgoraUserRepo
	UserBasicInfoRepo infrastructure.UserBasicInfoRepo
}

func (h *UserCreatedHandler) Handle(ctx context.Context, value []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	evtUserPayload := &upb.EvtUser{}
	err := proto.Unmarshal(value, evtUserPayload)
	if err != nil {
		h.Logger.Error(err.Error())
		return false, err
	}

	var agoraUser *models.AgoraUser
	switch evtUserPayload.Message.(type) {
	case *upb.EvtUser_CreateParent_:
		msg := evtUserPayload.GetCreateParent()
		agoraUser, err = mappers.CreateParentEvtToAgoraUserEnt(msg)
		if err != nil {
			return false, fmt.Errorf("err mappers.CreateParentEvtToAgoraUserEnt: %w", err)
		}
	case *upb.EvtUser_CreateStudent_:
		msg := evtUserPayload.GetCreateStudent()
		agoraUser, err = mappers.CreateStudentEvtToAgoraUserEnt(msg)
		if err != nil {
			return false, fmt.Errorf("err mappers.CreateStudentEvtToAgoraUserEnt: %w", err)
		}
	}
	if agoraUser == nil {
		return false, fmt.Errorf("cannot get agora user entity, nil value")
	}

	// TODO: Remove this code block after release, only for testing. Support limitation of Agora user test.
	// START TEMP BLOCK: Will get user info and if its name contain "AgoraTest" -> create agora user
	var user *models.UserBasicInfo
	err = try.Do(func(attempt int) (bool, error) {
		users, err := h.UserBasicInfoRepo.GetUsers(ctx, h.DB, []string{agoraUser.UserID.String})
		if err != nil {
			// Waiting for Kafka sync to users table on tom DB if needed
			if attempt < 5 {
				time.Sleep(1000 * time.Millisecond)
				return true, err
			}
			return false, err
		}

		// nolint
		if len(users) > 0 {
			user = users[0]
		} else if attempt < 5 {
			time.Sleep(1000 * time.Millisecond)
			return true, fmt.Errorf("not found user")
		} else {
			return false, fmt.Errorf("not found user")
		}

		return false, nil
	})
	if err != nil {
		return false, err
	}
	if !strings.Contains(user.GetName(), constants.UserNameConditionToCreateAgoraUser) || err != nil {
		return false, nil
	}
	// END TEMP BLOCK: End check condition to create Agora user.

	err = try.Do(func(attempt int) (bool, error) {
		createAgoraUserReq := mappers.AgoraUserToCreateUserReq(agoraUser)
		_, err := h.ChatVendorClient.CreateUser(createAgoraUserReq)
		if err != nil {
			// Re-try when cannot create Agora user
			if attempt < constants.TimesRetryChatVendorRequest {
				time.Sleep(1 * time.Second)
				return true, fmt.Errorf("cannot create Agora user: [%v]", err)
			}

			return false, fmt.Errorf("cannot create Agora user: [%v]", err)
		}

		return false, nil
	})

	if err != nil {
		agoraUserFailure, errConvert := mappers.AgoraUserEntToAgoraUserFailureEnt(agoraUser)
		if errConvert != nil {
			return false, fmt.Errorf("error create Agora user: [%v], error get Agora user failure entity: [%v]", err, errConvert)
		}
		errCreateAgoraFailureUser := h.AgoraUserRepo.CreateAgoraUserFailure(ctx, h.DB, agoraUserFailure)
		if errConvert != nil {
			return false, fmt.Errorf("error create Agora user on Agora server: [%v], error create Agora user failure in DB: [%v]", err, errCreateAgoraFailureUser)
		}
	}

	err = h.AgoraUserRepo.Create(ctx, h.DB, agoraUser)
	if err != nil {
		return false, fmt.Errorf("error create Agora user in DB: [%v]", err)
	}

	return false, nil
}

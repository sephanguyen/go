package consumers

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/agora_usermgmt/infrastructure/repositories"
	mock_chatvendor "github.com/manabie-com/backend/mock/golibs/chatvendor"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func TestHandleCreateUserEvt(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	mockAgoraUserRepo := &mock_repositories.MockAgoraUserRepo{}
	mockUserBasicInfoRepo := &mock_repositories.MockUserBasicInfoRepo{}
	mockChatVendorClient := mock_chatvendor.NewChatVendorClient(t)

	handler := UserCreatedHandler{
		DB:                mockDB,
		Logger:            *zap.NewExample(),
		ChatVendorClient:  mockChatVendorClient,
		AgoraUserRepo:     mockAgoraUserRepo,
		UserBasicInfoRepo: mockUserBasicInfoRepo,
	}

	t.Run("happy case create student", func(t *testing.T) {
		studentID := idutil.ULIDNow()
		agoraUserID := utils.GetAgoraUserID(studentID)
		payload := &upb.EvtUser{
			Message: &upb.EvtUser_CreateStudent_{
				CreateStudent: &upb.EvtUser_CreateStudent{
					StudentId: studentID,
				},
			},
		}
		bytePayload, _ := proto.Marshal(payload)

		mockUserBasicInfoRepo.On("GetUsers", mock.Anything, mock.Anything, []string{studentID}).Once().Return([]*models.UserBasicInfo{
			{
				UserID:    database.Text(studentID),
				FullName:  database.Text("AgoraTest"),
				FirstName: database.Text("AgoraTest"),
				LastName:  database.Text("AgoraTest"),
			},
		}, nil)

		mockChatVendorClient.On("CreateUser", &chatvendor_dto.CreateUserRequest{
			UserID:       studentID,
			VendorUserID: agoraUserID,
		}).Once().Return(&chatvendor_dto.CreateUserResponse{
			User: entities.User{
				UserID:       studentID,
				VendorUserID: agoraUserID,
			},
		}, nil)

		agoraUser := &models.AgoraUser{}
		database.AllNullEntity(agoraUser)
		agoraUser.UserID = database.Text(studentID)
		agoraUser.AgoraUserID = database.Text(agoraUserID)
		mockAgoraUserRepo.On("Create", mock.Anything, mock.Anything, agoraUser).Once().Return(nil)

		_, err := handler.Handle(context.Background(), bytePayload)
		assert.Nil(t, err)
	})

	t.Run("happy case create parent", func(t *testing.T) {
		parentID := idutil.ULIDNow()
		agoraUserID := utils.GetAgoraUserID(parentID)
		payload := &upb.EvtUser{
			Message: &upb.EvtUser_CreateParent_{
				CreateParent: &upb.EvtUser_CreateParent{
					ParentId: parentID,
				},
			},
		}
		bytePayload, _ := proto.Marshal(payload)

		mockUserBasicInfoRepo.On("GetUsers", mock.Anything, mock.Anything, []string{parentID}).Once().Return([]*models.UserBasicInfo{
			{
				UserID:    database.Text(parentID),
				FullName:  database.Text("AgoraTest"),
				FirstName: database.Text("AgoraTest"),
				LastName:  database.Text("AgoraTest"),
			},
		}, nil)

		mockChatVendorClient.On("CreateUser", &chatvendor_dto.CreateUserRequest{
			UserID:       parentID,
			VendorUserID: agoraUserID,
		}).Once().Return(&chatvendor_dto.CreateUserResponse{
			User: entities.User{
				UserID:       parentID,
				VendorUserID: agoraUserID,
			},
		}, nil)

		agoraUser := &models.AgoraUser{}
		database.AllNullEntity(agoraUser)
		agoraUser.UserID = database.Text(parentID)
		agoraUser.AgoraUserID = database.Text(agoraUserID)
		mockAgoraUserRepo.On("Create", mock.Anything, mock.Anything, agoraUser).Once().Return(nil)

		_, err := handler.Handle(context.Background(), bytePayload)
		assert.Nil(t, err)
	})
}

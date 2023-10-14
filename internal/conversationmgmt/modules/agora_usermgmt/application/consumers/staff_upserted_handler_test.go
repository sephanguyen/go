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

func TestHandleUpsertStaffEvt(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockAgoraUserRepo := &mock_repositories.MockAgoraUserRepo{}
	mockUserBasicInfoRepo := &mock_repositories.MockUserBasicInfoRepo{}
	mockChatVendorClient := mock_chatvendor.NewChatVendorClient(t)

	handler := StaffUpsertedHandler{
		DB:                mockDB,
		Logger:            *zap.NewExample(),
		ChatVendorClient:  mockChatVendorClient,
		AgoraUserRepo:     mockAgoraUserRepo,
		UserBasicInfoRepo: mockUserBasicInfoRepo,
	}

	t.Run("happy case create student", func(t *testing.T) {
		t.Parallel()
		staffID := idutil.ULIDNow()
		agoraUserID := utils.GetAgoraUserID(staffID)
		payload := &upb.EvtUpsertStaff{
			StaffId:      staffID,
			Type:         upb.EvtUpsertStaff_UPSERT_STAFF_TYPE_CREATE,
			UserGroupIds: []string{},
			LocationIds:  []string{},
		}
		bytePayload, _ := proto.Marshal(payload)

		mockUserBasicInfoRepo.On("GetUsers", mock.Anything, mock.Anything, []string{staffID}).Once().Return([]*models.UserBasicInfo{
			{
				UserID:    database.Text(staffID),
				FullName:  database.Text("AgoraTest"),
				FirstName: database.Text("AgoraTest"),
				LastName:  database.Text("AgoraTest"),
			},
		}, nil)

		mockChatVendorClient.On("CreateUser", &chatvendor_dto.CreateUserRequest{
			UserID:       staffID,
			VendorUserID: agoraUserID,
		}).Once().Return(&chatvendor_dto.CreateUserResponse{
			User: entities.User{
				UserID:       staffID,
				VendorUserID: agoraUserID,
			},
		}, nil)

		agoraUser := &models.AgoraUser{}
		database.AllNullEntity(agoraUser)
		agoraUser.UserID = database.Text(staffID)
		agoraUser.AgoraUserID = database.Text(agoraUserID)
		mockAgoraUserRepo.On("Create", mock.Anything, mock.Anything, agoraUser).Once().Return(nil)

		_, err := handler.Handle(context.Background(), bytePayload)
		assert.Nil(t, err)
	})
}

package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func Test_UpdateUserDeviceToken(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	jsm := &mock_nats.JetStreamManagement{}
	svc := &NotificationModifierService{
		DB:                  mockDB,
		UserDeviceTokenRepo: userDeviceTokenRepo,
		JSM:                 jsm,
		UserRepo:            userRepo,
	}

	type TestCase struct {
		Name         string
		Request      interface{}
		ExpcResponse interface{}
		ExpcErr      error
		Setup        func(ctx context.Context, this *TestCase)
	}

	testCases := []TestCase{
		{
			Name: "happy case",
			Request: &npb.UpdateUserDeviceTokenRequest{
				UserId:            idutil.ULIDNow(),
				DeviceToken:       idutil.ULIDNow(),
				AllowNotification: true,
			},
			ExpcResponse: &npb.UpdateUserDeviceTokenResponse{Successful: true},
			ExpcErr:      nil,
			Setup: func(ctx context.Context, this *TestCase) {
				req := this.Request.(*npb.UpdateUserDeviceTokenRequest)
				userName := "lastName firstName"
				userDeviceToken := &entities.UserDeviceToken{
					UserID:            database.Text(req.UserId),
					DeviceToken:       database.Text(req.DeviceToken),
					AllowNotification: database.Bool(req.AllowNotification),
				}
				userDeviceTokenRepo.On("UpsertUserDeviceToken", ctx, mockDB, userDeviceToken).Once().Return(nil)

				filter := repositories.NewFindUserFilter()
				_ = filter.UserIDs.Set([]string{req.UserId})
				users := []*entities.User{
					{
						UserID: database.Text(req.UserId),
						Name:   database.Text(userName),
					},
				}
				userRepo.On("FindUser", ctx, mockDB, filter, mock.Anything).Once().Return(users, map[string]*entities.User{}, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectUserDeviceTokenUpdated, mock.MatchedBy(func(msg []byte) bool {
					evt := &upb.EvtUserInfo{}
					_ = proto.Unmarshal(msg, evt)
					if evt.GetName() != userName {
						return false
					}
					return true
				})).Once().Return("", nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx, &testCase)
			resp, err := svc.UpdateUserDeviceToken(ctx, testCase.Request.(*npb.UpdateUserDeviceTokenRequest))
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpcResponse, resp)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}

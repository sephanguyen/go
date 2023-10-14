package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNotificationModifierService_DiscardNotification(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	questionnaireRepo := &mock_repositories.MockQuestionnaireRepo{}
	questionnaireQuestionRepo := &mock_repositories.MockQuestionnaireQuestionRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	infoNotificationAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}
	infoNotificationTagRepo := &mock_repositories.MockInfoNotificationTagRepo{}
	svc := &NotificationModifierService{
		DB:                             mockDB,
		InfoNotificationRepo:           infoNotificationRepo,
		InfoNotificationMsgRepo:        infoNotificationMsgRepo,
		InfoNotificationTagRepo:        infoNotificationTagRepo,
		InfoNotificationAccessPathRepo: infoNotificationAccessPathRepo,
		QuestionnaireRepo:              questionnaireRepo,
		QuestionnaireQuestionRepo:      questionnaireQuestionRepo,
	}
	statues := []string{cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()}
	testCases := []struct {
		Name  string
		Err   error
		Req   *npb.DiscardNotificationRequest
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req: &npb.DiscardNotificationRequest{
				NotificationId: "notification_id_1",
			},
			Setup: func(ctx context.Context) {

				es := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:    database.Text("notification_id_1"),
					NotificationMsgID: database.Text("notification_msg_id_1"),
					Status:            database.Text("NOTIFICATION_STATUS_DRAFT"),
					TargetGroups: database.JSONB(&entities.InfoNotificationTarget{
						UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{
							UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String(), cpb.UserGroup_USER_GROUP_STUDENT.String()},
						},
					}),
					Owner:           database.Int4(constant.ManabieSchool),
					QuestionnaireID: database.Text("questionnaire_id_1"),
				}}
				findNotificationFilter := repositories.NewFindNotificationFilter()
				findNotificationFilter.NotiIDs.Set([]string{"notification_id_1"})
				infoNotificationRepo.On("Find", ctx, svc.DB, findNotificationFilter).Once().Return(es, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				softDeleteNotificationTagFilter := repositories.NewSoftDeleteNotificationTagFilter()
				_ = softDeleteNotificationTagFilter.NotificationIDs.Set([]string{es[0].NotificationID.String})
				infoNotificationTagRepo.On("SoftDelete", ctx, mockTx, softDeleteNotificationTagFilter).Once().Return(nil)

				softDeleteNotificationAccessPathFilter := repositories.NewSoftDeleteNotificationAccessPathFilter()
				_ = softDeleteNotificationAccessPathFilter.NotificationIDs.Set([]string{es[0].NotificationID.String})
				infoNotificationAccessPathRepo.On("SoftDelete", ctx, mockTx, softDeleteNotificationAccessPathFilter).Once().Return(nil)

				questionnaireRepo.On("SoftDelete", ctx, mockTx, []string{"questionnaire_id_1"}).Once().Return(nil)
				questionnaireQuestionRepo.On("SoftDelete", ctx, mockTx, []string{"questionnaire_id_1"}).Once().Return(nil)
				infoNotificationRepo.On("DiscardNotification", ctx, mockTx, database.Text("notification_id_1"), database.TextArray(statues)).Once().Return(nil)
				infoNotificationMsgRepo.On("SoftDelete", ctx, mockTx, []string{es[0].NotificationMsgID.String}).Once().Return(nil)
			},
		},
		{
			Name: "sent notification",
			Err:  status.Error(codes.InvalidArgument, fmt.Errorf("the notification has been sent, you can no longer discard this notification").Error()),
			Req: &npb.DiscardNotificationRequest{
				NotificationId: "notification_id_1",
			},
			Setup: func(ctx context.Context) {
				es := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:    database.Text("notification_id_1"),
					NotificationMsgID: database.Text("notification_msg_id_1"),
					Status:            database.Text("NOTIFICATION_STATUS_SENT"),
					TargetGroups: database.JSONB(&entities.InfoNotificationTarget{
						UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{
							UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String(), cpb.UserGroup_USER_GROUP_STUDENT.String()},
						},
					}),
					Owner: database.Int4(constant.ManabieSchool),
				}}
				findNotificationFilter := repositories.NewFindNotificationFilter()
				findNotificationFilter.NotiIDs.Set([]string{"notification_id_1"})
				infoNotificationRepo.On("Find", ctx, mockDB, findNotificationFilter).Once().Return(es, nil)
			},
		},
		{
			Name: "deleted notification",
			Err:  status.Error(codes.InvalidArgument, fmt.Errorf("the notification has been deleted").Error()),
			Req: &npb.DiscardNotificationRequest{
				NotificationId: "notification_id_1",
			},
			Setup: func(ctx context.Context) {
				findNotificationFilter := repositories.NewFindNotificationFilter()
				findNotificationFilter.NotiIDs.Set([]string{"notification_id_1"})
				infoNotificationRepo.On("Find", ctx, mockDB, findNotificationFilter).Once().Return(entities.InfoNotifications{}, fmt.Errorf("InfoNotificationRepo.Find: can not find notification with id notification_id_1"))
				infoNotificationRepo.On("IsNotificationDeleted", ctx, mockDB, database.Text("notification_id_1")).Once().Return(true, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			testCase.Setup(ctx)
			_, err := svc.DiscardNotification(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, testCase.Err)
			} else {
				assert.Equal(t, testCase.Err.Error(), err.Error())
			}
		})
	}
}

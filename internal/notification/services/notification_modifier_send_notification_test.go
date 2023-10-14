package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	mock_infra "github.com/manabie-com/backend/mock/notification/infra"
	mock_metrics "github.com/manabie-com/backend/mock/notification/infra/metrics"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	mock_domain_services "github.com/manabie-com/backend/mock/notification/services/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNotificationModifierService_SendNotification(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	studentRepo := &mock_bob_repositories.MockStudentRepo{}
	studentParentRepo := &mock_bob_repositories.MockStudentParentRepo{}
	activityLogRepo := &mock_bob_repositories.MockActivityLogRepo{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	notificationAudienceRetrieverSvc := &mock_domain_services.MockAudienceRetrieverService{}
	notificationDataRetentionSvc := &mock_domain_services.MockDataRetentionService{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	ifntAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}

	mockNotificationMetric := &mock_metrics.NotificationMetrics{}
	mockNotificationMetric.On("RecordUserNotificationCreated", mock.Anything)

	pushNotificationService := &mock_infra.PushNotificationService{}

	studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}

	studentCoursesMap := make(map[string][]string)
	studentCoursesMap[studentIDs[0]] = []string{"course_id_1"}
	studentCoursesMap[studentIDs[1]] = []string{"course_id_2"}
	studentCoursesMap[studentIDs[2]] = []string{"course_id_3"}
	userID := "user-id"

	svc := &NotificationModifierService{
		DB:                             mockDB,
		InfoNotificationRepo:           infoNotificationRepo,
		InfoNotificationMsgRepo:        infoNotificationMsgRepo,
		UserNotificationRepo:           userInfoNotificationRepo,
		StudentRepo:                    studentRepo,
		StudentParentRepo:              studentParentRepo,
		ActivityLogRepo:                activityLogRepo,
		PushNotificationService:        pushNotificationService,
		NotificationMetrics:            mockNotificationMetric,
		UserDeviceTokenRepo:            userDeviceTokenRepo,
		NotificationAudienceRetriever:  notificationAudienceRetrieverSvc,
		DataRetentionService:           notificationDataRetentionSvc,
		LocationRepo:                   locationRepo,
		InfoNotificationAccessPathRepo: ifntAccessPathRepo,
	}
	notificationID := "notification-id-1"
	notificationMsgID := "notification-msg-id-1"
	individualIDs := []string{"individual_id_1", "individual_id_2", "individual_id_3"}
	userDeviceTokens := entities.UserDeviceTokens{
		{
			UserID:            database.Text("individual_id_1"),
			DeviceToken:       database.Text("device-token-1"),
			AllowNotification: database.Bool(true),
		},
		{
			UserID:            database.Text("individual_id_2"),
			DeviceToken:       database.Text("device-token-2"),
			AllowNotification: database.Bool(true),
		},
		{
			UserID:            database.Text("individual_id_3"),
			DeviceToken:       database.Text("device-token-3"),
			AllowNotification: database.Bool(true),
		},
	}
	testCases := []struct {
		Name    string
		Request *npb.SendNotificationRequest
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Request: &npb.SendNotificationRequest{
				NotificationId: notificationID,
			},
			Setup: func(ctx context.Context) {
				notis := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:    database.Text(notificationID),
					Status:            database.Text(cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()),
					NotificationMsgID: database.Text(notificationMsgID),
					ReceiverIDs:       database.TextArray(individualIDs),
					TargetGroups: database.JSONB(&entities.InfoNotificationTarget{
						UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{
							UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String(), cpb.UserGroup_USER_GROUP_STUDENT.String()},
						},
					}),
					Owner:         database.Int4(constant.ManabieSchool),
					CreatedUserID: database.Text(userID),
				}}

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				findNotificationFilter := repositories.NewFindNotificationFilter()
				// CanSendNotification
				findNotificationFilter.NotiIDs.Set([]string{notificationID})
				infoNotificationRepo.On("Find", ctx, mockDB, findNotificationFilter).Once().Return(notis, nil)

				notiMsges := entities.InfoNotificationMsgs{
					{
						NotificationMsgID: database.Text(notificationMsgID),
						Title:             database.Text("title of notification"),
					},
				}
				infoNotificationMsgRepo.On("GetByIDs", ctx, mockDB, database.TextArray([]string{notificationMsgID})).Once().Return(notiMsges, nil)

				userInfoNotificationRepo.On("Find", ctx, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)
				notificationAudienceRetrieverSvc.On("FindAudiences", mock.Anything, mockTx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				notificationDataRetentionSvc.On("AssignRetentionNameForUserNotification", mock.Anything, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)
				userInfoNotificationRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				uIDs := make([]string, 0)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				uIDs = append(uIDs, parentIDs...)
				uIDs = append(uIDs, studentIDs...)
				uIDs = append(uIDs, individualIDs...)

				infoNotificationRepo.On("UpdateNotification", ctx, mockTx, database.Text(notificationID), mock.Anything).Once().Return(nil)

				userDeviceTokenRepo.On("FindByUserIDs", mock.Anything, mockDB, database.TextArray(uIDs)).Once().Return(userDeviceTokens, nil)
				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, notis[0], notiMsges[0]).Once().Return(0, 0, nil)
				activityLogRepo.On("Create", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserID:       "user-id",
			ResourcePath: "resource_path",
		},
	})
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			_, err := svc.SendNotification(ctx, testCase.Request)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNotificationModifierService_SendNotificationToTargeWithoutSave(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	userID := idutil.ULIDNow()

	notification := utils.GenSampleNotification()
	notification.EditorId = userID
	notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
	}
	notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
	}

	notification.ScheduledAt = timestamppb.New(time.Now().Add(time.Hour))
	individualIDs := []string{"individual_id_1", "individual_id_2", "individual_id_3"}
	studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
	parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}

	userDeviceTokens := entities.UserDeviceTokens{
		{
			UserID:            database.Text("individual_id_1"),
			DeviceToken:       database.Text("device-token-1"),
			AllowNotification: database.Bool(true),
		},
		{
			UserID:            database.Text("individual_id_2"),
			DeviceToken:       database.Text("device-token-2"),
			AllowNotification: database.Bool(true),
		},
		{
			UserID:            database.Text("individual_id_3"),
			DeviceToken:       database.Text("device-token-3"),
			AllowNotification: database.Bool(true),
		},
	}
	notification.ReceiverIds = append(notification.ReceiverIds, individualIDs...)

	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	studentRepo := &mock_bob_repositories.MockStudentRepo{}
	studentParentRepo := &mock_bob_repositories.MockStudentParentRepo{}
	pushNotificationService := &mock_infra.PushNotificationService{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	notificationAudienceRetrieverSvc := &mock_domain_services.MockAudienceRetrieverService{}
	notificationDataRetentionrSvc := &mock_domain_services.MockDataRetentionService{}
	notificationUserRepo := &mock_repositories.MockUserRepo{}

	svc := &NotificationModifierService{
		DB:                            mockDB,
		InfoNotificationRepo:          infoNotificationRepo,
		InfoNotificationMsgRepo:       infoNotificationMsgRepo,
		UserNotificationRepo:          userInfoNotificationRepo,
		StudentRepo:                   studentRepo,
		StudentParentRepo:             studentParentRepo,
		PushNotificationService:       pushNotificationService,
		UserDeviceTokenRepo:           userDeviceTokenRepo,
		NotificationAudienceRetriever: notificationAudienceRetrieverSvc,
		DataRetentionService:          notificationDataRetentionrSvc,
		NotificationUserRepo:          notificationUserRepo,
	}

	testCases := []struct {
		Name         string
		Notification *cpb.Notification
		Err          error
		Setup        func(ctx context.Context)
	}{
		{
			Name: "happy case with generic user_id",
			Err:  nil,
			Notification: func() *cpb.Notification {
				ret := utils.GenSampleNotification()
				ret.GenericReceiverIds = studentIDs
				return ret
			}(),
			Setup: func(ctx context.Context) {
				uIDs := make([]string, 0)
				uIDs = append(uIDs, studentIDs...)
				users := []*entities.User{}
				for _, studentID := range studentIDs {
					users = append(users, &entities.User{
						UserID: database.Text(studentID),
					})
				}
				notificationAudienceRetrieverSvc.On("FindAudiences", mock.Anything, mockDB, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

				notificationUserRepo.On("FindUser", ctx, mockDB, mock.Anything).Once().Return(users, map[string]*entities.User{}, nil)
				userDeviceTokenRepo.On("FindByUserIDs", ctx, mockDB, database.TextArray(uIDs)).Once().Return(userDeviceTokens, nil)
				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, mock.Anything, mock.Anything).Once().Return(0, 0, nil)
			},
		},
		{
			Name:         "happy case with target group",
			Err:          nil,
			Notification: notification,
			Setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				notificationAudienceRetrieverSvc.On("FindAudiences", mock.Anything, mockDB, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

				uIDs := make([]string, 0)
				uIDs = append(uIDs, studentIDs...)
				uIDs = append(uIDs, parentIDs...)

				userDeviceTokenRepo.On("FindByUserIDs", ctx, mockDB, database.TextArray(uIDs)).Once().Return(userDeviceTokens, nil)
				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, mock.Anything, mock.Anything).Once().Return(0, 0, nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			// ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			err := svc.SendNotificationToTargetWithoutSave(ctx, testCase.Notification)
			assert.Nil(t, err)
		})
	}
}

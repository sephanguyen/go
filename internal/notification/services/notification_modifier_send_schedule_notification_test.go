package services

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
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

func TestNotificationModifierService_SendScheduledNotification(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	studentRepo := &mock_bob_repositories.MockStudentRepo{}
	studentParentRepo := &mock_bob_repositories.MockStudentParentRepo{}
	activityLogRepo := &mock_bob_repositories.MockActivityLogRepo{}
	organizationRepo := &mock_bob_repositories.MockOrganizationRepo{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	notificationInternalUserRepo := &mock_repositories.MockNotificationInternalUserRepo{}
	notificationAudienceRetrieverSvc := &mock_domain_services.MockAudienceRetrieverService{}
	notificationDataRetentionSvc := &mock_domain_services.MockDataRetentionService{}
	mockNotificationMetric := &mock_metrics.NotificationMetrics{}
	mockNotificationMetric.On("RecordUserNotificationCreated", mock.Anything)
	locationRepo := &mock_repositories.MockLocationRepo{}
	ifntAccessPathRepo := &mock_repositories.MockInfoNotificationAccessPathRepo{}

	pushNotificationService := &mock_infra.PushNotificationService{}

	svc := &NotificationModifierService{
		DB:                             mockDB,
		InfoNotificationRepo:           infoNotificationRepo,
		InfoNotificationMsgRepo:        infoNotificationMsgRepo,
		UserNotificationRepo:           userInfoNotificationRepo,
		StudentRepo:                    studentRepo,
		StudentParentRepo:              studentParentRepo,
		ActivityLogRepo:                activityLogRepo,
		OrganizationRepo:               organizationRepo,
		PushNotificationService:        pushNotificationService,
		NotificationMetrics:            mockNotificationMetric,
		UserDeviceTokenRepo:            userDeviceTokenRepo,
		NotificationAudienceRetriever:  notificationAudienceRetrieverSvc,
		NotificationInternalUserRepo:   notificationInternalUserRepo,
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
	locationIDs := []string{"loc-1", "loc-2", "loc-3"}
	userID := "user-id"
	mapLocationAccessPath := map[string]string{locationIDs[0]: locationIDs[0], locationIDs[1]: locationIDs[1], locationIDs[2]: locationIDs[2]}
	notificationPermissions := []string{
		consts.NotificationWritePermission,
		consts.NotificationOwnerPermission,
	}

	from := time.Now().Round(time.Minute)
	to := from.Add(time.Minute)
	testCases := []struct {
		Name    string
		Request *npb.SendScheduledNotificationRequest
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case full tenant",
			Err:  nil,
			Request: &npb.SendScheduledNotificationRequest{
				To:                     timestamppb.New(to),
				IsRunningForAllTenants: true,
			},
			Setup: func(ctx context.Context) {
				customClaims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						ResourcePath: strconv.Itoa(constant.ManabieSchool),
						UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
						UserID:       "internal-user-id",
					},
				}
				ctx = interceptors.ContextWithJWTClaims(ctx, customClaims)

				organizationRepo.On("GetOrganizations", mock.Anything, mock.Anything).Return([]string{fmt.Sprintf("%v", constant.ManabieSchool)}, nil)
				notificationInternalUserRepo.On("GetByOrgID", mock.Anything, mock.Anything, fmt.Sprint(constant.ManabieSchool)).Return(&entities.NotificationInternalUser{
					UserID: database.Text("internal-user-id"),
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				es := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:    database.Text(notificationID),
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

				findNotificationFilter := repositories.NewFindNotificationFilter()
				findNotificationFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				findNotificationFilter.ToScheduled = database.TimestamptzFromPb(timestamppb.New(to))
				findNotificationFilter.ResourcePath = database.Text(strconv.Itoa(constant.ManabieSchool))

				infoNotificationRepo.On("Find", mock.Anything, mockDB, findNotificationFilter).Once().Return(es, nil)
				ees := entities.InfoNotificationMsgs{
					{
						NotificationMsgID: database.Text(notificationMsgID),
						Title:             database.Text("title of notification"),
					},
				}
				infoNotificationMsgRepo.On("GetByIDs", mock.Anything, mockDB, database.TextArray([]string{notificationMsgID})).Once().Return(ees, nil)

				// upsert noti access paths
				locationRepo.On("GetGrantedLocationsByUserIDAndPermissions", ctx, mockTx, userID, notificationPermissions).Once().Return(locationIDs, mapLocationAccessPath, nil)
				notificationAccessPathsDelete := entities.InfoNotificationAccessPaths{
					{
						NotificationID: database.Text(notificationID),
						LocationID:     database.Text(locationIDs[0]),
						AccessPath:     database.Text(locationIDs[0]),
					},
				}
				ifntAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, notificationID, locationIDs).
					Once().Return(notificationAccessPathsDelete, nil)
				softDeleteNotificationAccessPathFilter := repositories.NewSoftDeleteNotificationAccessPathFilter()
				_ = softDeleteNotificationAccessPathFilter.NotificationIDs.Set([]string{notificationID})
				_ = softDeleteNotificationAccessPathFilter.LocationIDs.Set([]string{locationIDs[0]})
				ifntAccessPathRepo.On("SoftDelete", ctx, mockTx, softDeleteNotificationAccessPathFilter).
					Once().Return(nil)
				ifntAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				userInfoNotificationRepo.On("Find", mock.Anything, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)

				studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
				notificationAudienceRetrieverSvc.On("FindAudiences", mock.Anything, mockTx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				notificationDataRetentionSvc.On("AssignRetentionNameForUserNotification", mock.Anything, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)
				userInfoNotificationRepo.On("Upsert", mock.Anything, mockTx, mock.Anything).Once().Return(nil)

				uIDs := make([]string, 0)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				uIDs = append(uIDs, parentIDs...)
				uIDs = append(uIDs, studentIDs...)
				uIDs = append(uIDs, individualIDs...)

				infoNotificationRepo.On("UpdateNotification", mock.Anything, mockTx, database.Text(notificationID), mock.Anything).Once().Return(nil)

				userDeviceTokenRepo.On("FindByUserIDs", mock.Anything, mockDB, database.TextArray(uIDs)).Once().Return(userDeviceTokens, nil)
				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, es[0], ees[0]).Once().Return(0, 0, nil)
				activityLogRepo.On("Create", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "happy case specific tenant",
			Err:  nil,
			Request: &npb.SendScheduledNotificationRequest{
				To:                     timestamppb.New(to),
				IsRunningForAllTenants: false,
				TenantIds:              []string{fmt.Sprintf("%v", constant.ManabieSchool)},
			},
			Setup: func(ctx context.Context) {
				customClaims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						ResourcePath: strconv.Itoa(constant.ManabieSchool),
						UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
						UserID:       "internal-user-id",
					},
				}
				ctx = interceptors.ContextWithJWTClaims(ctx, customClaims)

				notificationInternalUserRepo.On("GetByOrgID", mock.Anything, mock.Anything, fmt.Sprint(constant.ManabieSchool)).Return(&entities.NotificationInternalUser{
					UserID: database.Text("internal-user-id"),
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				es := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:    database.Text(notificationID),
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

				findNotificationFilter := repositories.NewFindNotificationFilter()
				findNotificationFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				findNotificationFilter.ToScheduled = database.TimestamptzFromPb(timestamppb.New(to))
				findNotificationFilter.ResourcePath = database.Text(strconv.Itoa(constant.ManabieSchool))

				infoNotificationRepo.On("Find", mock.Anything, mockDB, findNotificationFilter).Once().Return(es, nil)
				ees := entities.InfoNotificationMsgs{
					{
						NotificationMsgID: database.Text(notificationMsgID),
						Title:             database.Text("title of notification"),
					},
				}
				infoNotificationMsgRepo.On("GetByIDs", mock.Anything, mockDB, database.TextArray([]string{notificationMsgID})).Once().Return(ees, nil)

				// upsert noti access paths
				locationRepo.On("GetGrantedLocationsByUserIDAndPermissions", ctx, mockTx, userID, notificationPermissions).Once().Return(locationIDs, mapLocationAccessPath, nil)
				notificationAccessPathsDelete := entities.InfoNotificationAccessPaths{
					{
						NotificationID: database.Text(notificationID),
						LocationID:     database.Text(locationIDs[0]),
						AccessPath:     database.Text(locationIDs[0]),
					},
				}
				ifntAccessPathRepo.On("GetByNotificationIDAndNotInLocationIDs", ctx, mockTx, notificationID, locationIDs).
					Once().Return(notificationAccessPathsDelete, nil)
				softDeleteNotificationAccessPathFilter := repositories.NewSoftDeleteNotificationAccessPathFilter()
				_ = softDeleteNotificationAccessPathFilter.NotificationIDs.Set([]string{notificationID})
				_ = softDeleteNotificationAccessPathFilter.LocationIDs.Set([]string{locationIDs[0]})
				ifntAccessPathRepo.On("SoftDelete", ctx, mockTx, softDeleteNotificationAccessPathFilter).
					Once().Return(nil)
				ifntAccessPathRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				userInfoNotificationRepo.On("Find", mock.Anything, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)

				studentIDs := []string{"student_id_1", "student_id_2", "student_id_3"}
				notificationAudienceRetrieverSvc.On("FindAudiences", mock.Anything, mockTx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				notificationDataRetentionSvc.On("AssignRetentionNameForUserNotification", mock.Anything, mockTx, mock.Anything).Once().Return(entities.UserInfoNotifications{}, nil)
				userInfoNotificationRepo.On("Upsert", mock.Anything, mockTx, mock.Anything).Once().Return(nil)

				uIDs := make([]string, 0)
				parentIDs := []string{"parent_id_1", "parent_id_2", "parent_id_3"}
				uIDs = append(uIDs, parentIDs...)
				uIDs = append(uIDs, studentIDs...)
				uIDs = append(uIDs, individualIDs...)

				infoNotificationRepo.On("UpdateNotification", mock.Anything, mockTx, database.Text(notificationID), mock.Anything).Once().Return(nil)

				userDeviceTokenRepo.On("FindByUserIDs", mock.Anything, mockDB, database.TextArray(uIDs)).Once().Return(userDeviceTokens, nil)
				pushNotificationService.On("PushNotificationForUser", ctx, userDeviceTokens, es[0], ees[0]).Once().Return(0, 0, nil)
				activityLogRepo.On("Create", mock.Anything, mockDB, mock.Anything).Once().Return(nil)
			},
		},
	}
	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			_, err := svc.SendScheduledNotification(ctx, testCase.Request)
			if testCase.Err != nil {
				assert.ErrorIs(t, err, testCase.Err)
			}
		})
	}
}

package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetNotificationsByFilter(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	infoNotiRepo := &mock_repositories.MockInfoNotificationRepo{}
	infoNotificationMsgRepo := &mock_repositories.MockInfoNotificationMsgRepo{}
	infoNotificationTagRepo := &mock_repositories.MockInfoNotificationTagRepo{}
	notificationLocationFilterRepo := &mock_repositories.MockNotificationLocationFilterRepo{}
	notificationCourseFilterRepo := &mock_repositories.MockNotificationCourseFilterRepo{}
	notificationClassFilterRepo := &mock_repositories.MockNotificationClassFilterRepo{}
	userNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}

	svc := &NotificationReaderService{
		DB:                             db,
		InfoNotificationRepo:           infoNotiRepo,
		InfoNotificationMsgRepo:        infoNotificationMsgRepo,
		InfoNotificationTagRepo:        infoNotificationTagRepo,
		NotificationLocationFilterRepo: notificationLocationFilterRepo,
		NotificationCourseFilterRepo:   notificationCourseFilterRepo,
		NotificationClassFilterRepo:    notificationClassFilterRepo,
		UserInfoNotificationRepo:       userNotificationRepo,
	}

	mockFromTime := timestamppb.Now()
	mockToTime := timestamppb.Now()

	testCases := []struct {
		Name  string
		Req   *npb.GetNotificationsByFilterRequest
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case, no filter",
			Req: &npb.GetNotificationsByFilterRequest{
				Keyword: "",
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			Setup: func(ctx context.Context) {
				notificationsFilter := repositories.NewFindNotificationFilter()
				notificationsFilter.Limit.Set(100)
				notificationsFilter.Offset.Set(0)
				notificationsFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				notificationsFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())

				infoNotiRepo.On("Find", ctx, db, notificationsFilter).Once().Return(entities.InfoNotifications{}, nil)

				infoNotificationMsgRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]*entities.InfoNotificationMsg), nil)
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]entities.InfoNotificationsTags), nil)

				countNotificationsForStatusFilter := repositories.NewFindNotificationFilter()
				countNotificationsForStatusFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())

				infoNotiRepo.On("CountTotalNotificationForStatus", ctx, db, countNotificationsForStatusFilter).Once().Return(make(map[string]uint32), nil)
			},
		},
		{
			Name: "happy case, keyword",
			Req: &npb.GetNotificationsByFilterRequest{
				Keyword: "keyword",
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			Setup: func(ctx context.Context) {
				notificationsFilter := repositories.NewFindNotificationFilter()
				notificationsFilter.Limit.Set(100)
				notificationsFilter.Offset.Set(0)
				notificationsFilter.NotificationMsgIDs.Set([]string{"notification-msg-id"})
				notificationsFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				notificationsFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())

				infoNotificationMsgRepo.On("GetIDsByTitle", ctx, db, database.Text("keyword")).Once().Return([]string{"notification-msg-id"}, nil)

				infoNotiRepo.On("Find", ctx, db, notificationsFilter).Once().Return(entities.InfoNotifications{}, nil)

				infoNotificationMsgRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]*entities.InfoNotificationMsg), nil)
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]entities.InfoNotificationsTags), nil)

				countNotificationsForStatusFilter := repositories.NewFindNotificationFilter()
				countNotificationsForStatusFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
				countNotificationsForStatusFilter.NotificationMsgIDs.Set([]string{"notification-msg-id"})

				infoNotiRepo.On("CountTotalNotificationForStatus", ctx, db, countNotificationsForStatusFilter).Once().Return(make(map[string]uint32), nil)
			},
		},
		{
			Name: "happy case, composer filter, from time filter, to time filter, tag filter, location filter, course filter, class filter, status filter, is questionnaire fully submitted filter",
			Req: &npb.GetNotificationsByFilterRequest{
				Keyword: "",
				TagIds:  []string{"tag-id"},
				TargetGroup: &cpb.NotificationTargetGroup{
					LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
						LocationIds: []string{"location-id"},
					},
					CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
						CourseIds: []string{"course-id"},
					},
					ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
						ClassIds: []string{"class-id"},
					},
				},
				ComposerIds:                   []string{"composer-id"},
				SentFrom:                      mockFromTime,
				SentTo:                        mockToTime,
				Status:                        cpb.NotificationStatus_NOTIFICATION_STATUS_NONE,
				IsQuestionnaireFullySubmitted: true,
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			Setup: func(ctx context.Context) {
				notificationsFilter := repositories.NewFindNotificationFilter()
				notificationsFilter.Limit.Set(100)
				notificationsFilter.Offset.Set(0)
				notificationsFilter.NotiIDs.Set([]string{"notification-id"})
				notificationsFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				notificationsFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
				notificationsFilter.EditorIDs.Set([]string{"composer-id"})
				notificationsFilter.FromSent = database.TimestamptzFromPb(mockFromTime)
				notificationsFilter.ToSent = database.TimestamptzFromPb(mockToTime)

				infoNotificationTagRepo.On("GetNotificationIDsByTagIDs", ctx, db, database.TextArray([]string{"tag-id"})).Once().Return([]string{"notification-id"}, nil)
				notificationLocationFilterRepo.On("GetNotificationIDsByLocationIDs", ctx, db, database.TextArray([]string{"notification-id"}), database.TextArray([]string{"location-id"})).Once().Return([]string{"notification-id"}, nil)
				notificationCourseFilterRepo.On("GetNotificationIDsByCourseIDs", ctx, db, database.TextArray([]string{"notification-id"}), database.TextArray([]string{"course-id"})).Once().Return([]string{"notification-id"}, nil)
				notificationClassFilterRepo.On("GetNotificationIDsByClassIDs", ctx, db, database.TextArray([]string{"notification-id"}), database.TextArray([]string{"class-id"})).Once().Return([]string{"notification-id"}, nil)
				userNotificationRepo.On("GetNotificationIDWithFullyQnStatus", ctx, db, database.TextArray([]string{"notification-id"}), database.Text("USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED")).Once().Return([]string{"notification-id"}, nil)

				infoNotiRepo.On("Find", ctx, db, notificationsFilter).Once().Return(entities.InfoNotifications{}, nil)

				infoNotificationMsgRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]*entities.InfoNotificationMsg), nil)
				infoNotificationTagRepo.On("GetByNotificationIDs", ctx, db, mock.Anything).Once().Return(make(map[string]entities.InfoNotificationsTags), nil)

				countNotificationsForStatusFilter := repositories.NewFindNotificationFilter()
				countNotificationsForStatusFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
				countNotificationsForStatusFilter.NotiIDs.Set([]string{"notification-id"})
				countNotificationsForStatusFilter.EditorIDs.Set([]string{"composer-id"})
				countNotificationsForStatusFilter.FromSent = database.TimestamptzFromPb(mockFromTime)
				countNotificationsForStatusFilter.ToSent = database.TimestamptzFromPb(mockToTime)

				infoNotiRepo.On("CountTotalNotificationForStatus", ctx, db, countNotificationsForStatusFilter).Once().Return(make(map[string]uint32), nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, mock.Anything)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.GetNotificationsByFilter(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}

package services

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	mock_domain_services "github.com/manabie-com/backend/mock/notification/services/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationReaderService_RetrieveDraftAudience(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	notificationAudienceRetrieverSvc := &mock_domain_services.MockAudienceRetrieverService{}
	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	svc := &NotificationReaderService{
		DB:                            db,
		NotificationAudienceRetriever: notificationAudienceRetrieverSvc,
		InfoNotificationRepo:          infoNotificationRepo,
	}
	audiences := []*entities.Audience{}
	audiencesPb := []*npb.RetrieveDraftAudienceResponse_Audience{}
	for i := 0; i < 10; i++ {
		userGroup := cpb.UserGroup_USER_GROUP_STUDENT.String()
		if i < 5 {
			userGroup = cpb.UserGroup_USER_GROUP_PARENT.String()
		}
		audiences = append(audiences, &entities.Audience{
			UserID:    database.Text(idutil.ULIDNow()),
			Name:      database.Text(idutil.ULIDNow()),
			Email:     database.Text(idutil.ULIDNow()),
			UserGroup: database.Text(userGroup),
		})
	}
	audiencesPb = mappers.NotificationDraftAudiencesToPb(audiences)
	notificationID := "notification-id"
	userEditor := "user-id-editor"
	testCases := []struct {
		Name  string
		Req   *npb.RetrieveDraftAudienceRequest
		Res   *npb.RetrieveDraftAudienceResponse
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: &npb.RetrieveDraftAudienceRequest{
				NotificationId: notificationID,
				Paging:         &cpb.Paging{Limit: 0, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0}},
			},
			Res: &npb.RetrieveDraftAudienceResponse{
				Audiences:  audiencesPb,
				TotalItems: uint32(len(audiencesPb)),
				NextPage: &cpb.Paging{
					Limit:  100,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(len(audiencesPb))},
				},
				PreviousPage: &cpb.Paging{
					Limit:  100,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
				},
			},
			Setup: func(ctx context.Context) {
				customClaims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						ResourcePath: strconv.Itoa(constant.ManabieSchool),
						UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
						UserID:       userEditor,
					},
				}
				ctxEditor := interceptors.ContextWithJWTClaims(context.Background(), customClaims)
				filter := repositories.NewFindNotificationFilter()
				_ = filter.NotiIDs.Set([]string{notificationID})
				_ = filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
				noti := utils.GenNotificationEntity()
				noti.NotificationID.Set(notificationID)
				noti.EditorID.Set(userEditor)

				infoNotificationRepo.On("Find", ctx, db, filter).Once().Return(entities.InfoNotifications([]*entities.InfoNotification{&noti}), nil)

				notificationAudienceRetrieverSvc.On("FindDraftAudiencesWithPaging", ctxEditor, mock.Anything,
					notificationID,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(audiences, uint32(len(audiences)), nil)
			},
		},
	}

	for _, testCase := range testCases {
		customClaims := &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: strconv.Itoa(constant.ManabieSchool),
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				UserID:       userEditor,
			},
		}
		ctx := interceptors.ContextWithJWTClaims(context.Background(), customClaims)
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			res, err := svc.RetrieveDraftAudience(ctx, testCase.Req)
			fmt.Printf("\n%v\n", err)
			assert.Nil(t, err)
			assert.Equal(t, testCase.Res, res)
		})
	}
}

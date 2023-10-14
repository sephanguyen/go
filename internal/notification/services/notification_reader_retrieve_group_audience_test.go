package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_domain_services "github.com/manabie-com/backend/mock/notification/services/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationReaderService_RetrieveGroupAudience(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	notificationAudienceRetrieverSvc := &mock_domain_services.MockAudienceRetrieverService{}
	svc := &NotificationReaderService{
		DB:                            db,
		NotificationAudienceRetriever: notificationAudienceRetrieverSvc,
	}
	audiences := []*entities.Audience{}
	audiencesPb := []*npb.RetrieveGroupAudienceResponse_Audience{}
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
	audiencesPb = mappers.NotificationGroupAudiencesToPb(audiences)

	testCases := []struct {
		Name  string
		Req   *npb.RetrieveGroupAudienceRequest
		Res   *npb.RetrieveGroupAudienceResponse
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: &npb.RetrieveGroupAudienceRequest{
				Keyword: "",
				TargetGroup: &cpb.NotificationTargetGroup{
					CourseFilter:    &cpb.NotificationTargetGroup_CourseFilter{},
					GradeFilter:     &cpb.NotificationTargetGroup_GradeFilter{},
					LocationFilter:  &cpb.NotificationTargetGroup_LocationFilter{},
					ClassFilter:     &cpb.NotificationTargetGroup_ClassFilter{},
					SchoolFilter:    &cpb.NotificationTargetGroup_SchoolFilter{},
					UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{},
				},
				Paging: &cpb.Paging{Limit: 0, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0}},
			},
			Res: &npb.RetrieveGroupAudienceResponse{
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
				notificationAudienceRetrieverSvc.On("FindGroupAudiencesWithPaging", ctx, mock.Anything,
					"",
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
		ctx := context.Background()
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			res, err := svc.RetrieveGroupAudience(ctx, testCase.Req)
			assert.Nil(t, err)
			assert.Equal(t, testCase.Res, res)
		})
	}
}

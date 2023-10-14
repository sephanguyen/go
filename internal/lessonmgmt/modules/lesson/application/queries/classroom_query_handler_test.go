package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClassroomQueryHandler_RetrieveClassroomsByLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	lessonClassroomRepo := new(mock_repositories.MockLessonClassroomRepo)

	h := &ClassroomQueryHandler{
		WrapperConnection:   wrapperConnection,
		ClassroomRepo:       classroomRepo,
		LessonClassroomRepo: lessonClassroomRepo,
	}
	timeLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	timezone := "Asia/Ho_Chi_Minh"

	testCases := []TestCase{
		{
			name: "happy case: get classrooms by locations without checking classroom status",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &lpb.RetrieveClassroomsByLocationIDRequest{
				LocationId: "location-1",
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				TimeZone: timezone,
			},
			result: []*domain.Classroom{
				{
					ClassroomID:     "classroom-1",
					Name:            "name 1",
					LocationID:      "location-1",
					RoomArea:        "floor 1",
					SeatCapacity:    20,
					ClassroomStatus: domain.Available,
				},
				{
					ClassroomID:     "classroom-2",
					Name:            "name 2",
					LocationID:      "location-1",
					RoomArea:        "floor 1",
					SeatCapacity:    40,
					ClassroomStatus: domain.InUsed,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classroomRepo.On("RetrieveClassroomsByLocationID", ctx, db, mock.Anything).Once().Return([]*domain.Classroom{
					{
						ClassroomID:     "classroom-1",
						Name:            "name 1",
						LocationID:      "location-1",
						RoomArea:        "floor 1",
						SeatCapacity:    20,
						ClassroomStatus: domain.Available,
					},
					{
						ClassroomID:     "classroom-2",
						Name:            "name 2",
						LocationID:      "location-1",
						RoomArea:        "floor 1",
						SeatCapacity:    40,
						ClassroomStatus: domain.InUsed,
					},
				}, nil)
			},
		},
		{
			name: "happy case - get fully classrooms info by locations",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &lpb.RetrieveClassroomsByLocationIDRequest{
				LocationId: "location-1",
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				TimeZone:  "Asia/Ho_Chi_Minh",
				StartTime: timestamppb.New(time.Date(2023, 01, 01, 0, 0, 0, 0, timeLocation)),
				EndTime:   timestamppb.New(time.Date(2023, 01, 01, 23, 59, 0, 0, timeLocation)),
			},
			result: []*domain.Classroom{
				{
					ClassroomID:     "classroom-1",
					Name:            "name 1",
					LocationID:      "location-1",
					RoomArea:        "floor 1",
					SeatCapacity:    20,
					ClassroomStatus: domain.Available,
				},
				{
					ClassroomID:     "classroom-2",
					Name:            "name 2",
					LocationID:      "location-1",
					RoomArea:        "floor 1",
					SeatCapacity:    40,
					ClassroomStatus: domain.InUsed,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classroomRepo.On("RetrieveClassroomsByLocationID", ctx, db, mock.Anything).Once().Return([]*domain.Classroom{
					{
						ClassroomID:     "classroom-1",
						Name:            "name 1",
						LocationID:      "location-1",
						RoomArea:        "floor 1",
						SeatCapacity:    20,
						ClassroomStatus: domain.Available,
					},
					{
						ClassroomID:     "classroom-2",
						Name:            "name 2",
						LocationID:      "location-1",
						RoomArea:        "floor 1",
						SeatCapacity:    40,
						ClassroomStatus: domain.InUsed,
					},
				}, nil)
				lessonClassroomRepo.On("GetOccupiedClassroomByTime", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&domain.LessonClassrooms{
					{
						ClassroomID: "classroom-2",
					},
				}, nil)
			},
		},
		{
			name:     "unsuccess",
			ctx:      interceptors.ContextWithUserID(ctx, "id"),
			payloads: &lpb.RetrieveClassroomsByLocationIDRequest{},
			result:   []*domain.Classroom{},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classroomRepo.On("RetrieveClassroomsByLocationID", ctx, db, mock.Anything).Once().
					Return(nil, fmt.Errorf("errSubString"))
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := h.RetrieveClassroomsByLocationID(testCase.ctx, testCase.payloads.(*lpb.RetrieveClassroomsByLocationIDRequest))
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.result, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

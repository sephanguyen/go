package queries

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStudentReallocateQueryHandler_RetrieveStudentPendingReallocate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	studentSubscriptionRepo := new(mock_repositories.MockStudentSubscriptionRepo)

	testCases := []struct {
		name   string
		setup  func(context.Context)
		hasErr bool
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.On("RetrieveStudentPendingReallocate", mock.Anything, mock.Anything, domain.RetrieveStudentPendingReallocateDto{
					Limit: 5,
				}).Once().Return([]*domain.ReallocateStudent{
					{
						StudentId:        "student-1",
						OriginalLessonID: "lesson-1",
						CourseID:         "course-1",
						LocationID:       "location-1",
						GradeID:          "grade-1",
						StartAt:          time.Now(),
						EndAt:            time.Now(),
					},
					{
						StudentId:        "student-2",
						OriginalLessonID: "lesson-1",
						CourseID:         "course-1",
						LocationID:       "location-1",
						GradeID:          "grade-1",
						StartAt:          time.Now(),
						EndAt:            time.Now(),
					},
				}, uint32(1), nil)
			},
		},
	}

	handler := StudentReallocateQueryHandler{
		WrapperConnection:       wrapperConnection,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		Env:                     "local",
		UnleashClientIns:        mockUnleashClient,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := handler.RetrieveStudentPendingReallocate(ctx, StudentReallocateRequest{
				Paging: support.Paging[int]{
					Limit: 5,
				},
			})
			if err != nil {
				require.True(t, tc.hasErr)
			} else {
				require.False(t, tc.hasErr)
				require.NotNil(t, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}

}

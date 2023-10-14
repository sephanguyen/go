package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"

	"google.golang.org/grpc/status"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignStudyPlanTaskModifierService_makeCourseStudyPlanUpdateData(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	s := &AssignStudyPlanTaskModifierService{
		DB:                db,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	studyPlanItems := []*pb.StudyPlanItem{
		{
			StudyPlanItemId: "study-plan-item-id-1",
		},
		{
			StudyPlanItemId: "study-plan-item-id-2",
		},
		{
			StudyPlanItemId: "study-plan-item-id-3",
		},
	}
	scheduleStudyPlanItems := []*pb.ScheduleStudyPlan{
		{
			StudyPlanItemId: "study-plan-item-id-1",
		},
		{
			StudyPlanItemId: "study-plan-item-id-2",
		},
		{
			StudyPlanItemId: "study-plan-item-id-3",
		},
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindByIDs", mock.Anything, db, mock.Anything).Once().Return(
					[]*entities.StudyPlanItem{
						{
							StudyPlanID: database.Text("study-plan-id-2"),
							CompletedAt: database.Timestamptz(time.Now()),
						},
					},
					nil,
				)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, _, _, err := s.makeCourseStudyPlanUpdateData(ctx, "course-study-plan-id", studyPlanItems, scheduleStudyPlanItems)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

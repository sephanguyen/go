package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities(t *testing.T) {
	t.Parallel()

	studentEventLogRepo := new(mock_repositories.MockStudentEventLogRepo)
	mockDB := &mock_database.Ext{}

	chunksInput := make([]*repositories.StudyPlanItemIdentity, 100)
	for i := 0; i < 100; i++ {
		chunksInput[i] = &repositories.StudyPlanItemIdentity{
			StudentID:          database.Text(fmt.Sprintf("student_id_%d", i)),
			StudyPlanID:        database.Text(fmt.Sprintf("study_plan_id_%d", i)),
			LearningMaterialID: database.Text(fmt.Sprintf("learning_material_id_%d", i)),
		}
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  context.Background(),
			req: []*repositories.StudyPlanItemIdentity{
				{
					StudentID:          database.Text("student_id"),
					StudyPlanID:        database.Text("study_plan_id"),
					LearningMaterialID: database.Text("learning_material_id"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "chunk case",
			ctx:         context.Background(),
			req:         chunksInput,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "repo err case",
			ctx:  context.Background(),
			req: []*repositories.StudyPlanItemIdentity{
				{
					StudentID:          database.Text("student_id"),
					StudyPlanID:        database.Text("study_plan_id"),
					LearningMaterialID: database.Text("learning_material_id"),
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities: %v", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.([]*repositories.StudyPlanItemIdentity)
			_, err := retrieveStudentEventLogsConcurrentlyByStudyPlanItemIdentities(testCase.ctx, mockDB, req, studentEventLogRepo)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

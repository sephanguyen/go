package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentSessionService_GetAssessmentSessionsByCourseAndLM(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		assessmentSessionRepo       = new(mock_repositories.MockAssessmentSessionRepo)
		assessmentRepo              = new(mock_repositories.MockAssessmentRepo)
		userRepo                    = new(mock_repositories.MockUserRepo)
		courseStudentAccessPathRepo = new(mock_repositories.MockCourseStudentAccessPathRepo)
		db                          = new(mock_database.Ext)
	)

	s := AssessmentSessionService{
		DB:                          db,
		AssessmentSessionRepo:       assessmentSessionRepo,
		AssessmentRepo:              assessmentRepo,
		UserRepo:                    userRepo,
		CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
	}

	assessmentEntity := &entities.Assessment{
		ID: database.Text("assessment-1"),
	}

	assessmentSessionEntity1 := &entities.AssessmentSession{
		SessionID:    database.Text("session-1"),
		AssessmentID: database.Text("assessment-1"),
		UserID:       database.Text("user-1"),
	}
	user1 := &entities.User{
		UserID: database.Text("user-1"),
		Name:   database.Text("name-1"),
	}

	listAssessmentE := []*entities.Assessment{assessmentEntity}
	listAssessmentSessionE := []*entities.AssessmentSession{assessmentSessionEntity1}
	listUsersE := []*entities.User{user1}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &sspb.GetAssessmentSessionsByCourseAndLMRequest{
				CourseId:           []string{"course-1"},
				LearningMaterialId: []string{"lm-1"},
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &sspb.GetAssessmentSessionsByCourseAndLMResponse{
				AssessmentSessions: []*sspb.GetAssessmentSessionsByCourseAndLMResponse_AssessmentSession{
					{
						SessionId:    "session-1",
						AssessmentId: "assessment-1",
						UserId:       "user-1",
						UserName:     "name-1",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				TotalItems: 1,
			},
			setup: func(ctx context.Context) {
				assessmentRepo.On("GetAssessmentByCourseAndLearningMaterial", ctx, db, mock.Anything, mock.Anything).
					Return(listAssessmentE, nil).Once()
				assessmentSessionRepo.On("GetAssessmentSessionByAssessmentIDs", ctx, db, mock.Anything).
					Return(listAssessmentSessionE, nil).Once()
				userRepo.On("GetUsersByIDs", ctx, db, mock.Anything).
					Return(listUsersE, nil).Once()
				assessmentSessionRepo.On("CountByAssessment", ctx, db, mock.Anything).
					Return(int32(1), nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetAssessmentSessionsByCourseAndLM(testCase.ctx, testCase.req.(*sspb.GetAssessmentSessionsByCourseAndLMRequest))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

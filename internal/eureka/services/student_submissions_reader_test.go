package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentSubmissionReader_RetrieveStudentSubmissionHistoryByLoIDs(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}

	studentEventLogRepo := &mock_repositories.MockStudentEventLogRepo{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}

	svc := NewStudentSubmissionReaderService(mockDB)
	svc.QuizSetRepo = quizSetRepo
	svc.StudentEventLogRepo = studentEventLogRepo

	testCases := []TestCase{
		{
			name:         "StudentEventLogRepo.LogsQuestionSubmitionByLO error",
			req:          &pb.RetrieveStudentSubmissionHistoryByLoIDsRequest{},
			expectedErr:  status.Errorf(codes.Internal, "StudentEventLogRepo.LogsQuestionSubmitionByLO %v", "StudentEventLogRepo.LogsQuestionSubmitionByLO error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("LogsQuestionSubmitionByLO", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("StudentEventLogRepo.LogsQuestionSubmitionByLO error"))
			},
		},
		{
			name:         "QuizSetRepo.CountQuizOnLO error",
			req:          &pb.RetrieveStudentSubmissionHistoryByLoIDsRequest{},
			expectedErr:  status.Errorf(codes.Internal, "QuizSetRepo.CountQuizOnLO %v", "QuizSetRepo.CountQuizOnLO error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("LogsQuestionSubmitionByLO", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string][]*repositories.QuestionSubmissionResult{}, nil)
				quizSetRepo.On("CountQuizOnLO", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("QuizSetRepo.CountQuizOnLO error"))
			},
		},
		{
			name: "happy case",
			req: &pb.RetrieveStudentSubmissionHistoryByLoIDsRequest{
				LoIds: []string{"lo_id"},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveStudentSubmissionHistoryByLoIDsResponse{
				Submissions: []*pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory{
					{
						LoId:          "lo_id",
						TotalQuestion: 5,
						Results: []*pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult{
							{
								QuestionId: "question-id",
								Correct:    true,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("LogsQuestionSubmitionByLO", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string][]*repositories.QuestionSubmissionResult{
					"lo_id": []*repositories.QuestionSubmissionResult{
						{
							QuestionID: "question-id",
							Correct:    true,
						},
					},
				}, nil)
				quizSetRepo.On("CountQuizOnLO", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]int32{
					"lo_id": 5,
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx == nil {
				tt.ctx = context.Background()
			}
			tt.setup(tt.ctx)
			resp, err := svc.RetrieveStudentSubmissionHistoryByLoIDs(tt.ctx, tt.req.(*pb.RetrieveStudentSubmissionHistoryByLoIDsRequest))
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})

	}
}

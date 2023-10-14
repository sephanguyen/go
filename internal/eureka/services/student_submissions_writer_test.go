package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentSubmissionWriterService_DeleteStudentSubmission(t *testing.T) {
	t.Parallel()
	submissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	latestSubmissionRepo := &mock_repositories.MockStudentLatestSubmissionRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	svc := &StudentSubmissionWriterService{
		DB:                          mockDB,
		StudentSubmissionRepo:       submissionRepo,
		StudentLatestSubmissionRepo: latestSubmissionRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
	}

	validReq := &pb.DeleteStudentSubmissionRequest{
		StudentSubmissionId: "student-submission-id",
	}

	testCases := []TestCase{
		{
			name:        "error no rows find student by class IDs",
			req:         &pb.DeleteStudentSubmissionRequest{},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission: No StudentSubmissionID").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "error DeleteStudentSubmission1",
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission1: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				submissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "error DeleteStudentSubmission2",
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission2: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentSubmission := &entities.StudentSubmission{
					StudyPlanItemID: database.Text("study-plan-item-id"),
				}
				submissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentSubmission, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				submissionRepo.On("DeleteByStudyPlanItemIDs", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "error DeleteStudentSubmission3",
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission3: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentSubmission := &entities.StudentSubmission{
					StudyPlanItemID: database.Text("study-plan-item-id"),
				}
				submissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentSubmission, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				submissionRepo.On("DeleteByStudyPlanItemIDs", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(nil)
				latestSubmissionRepo.On("DeleteByStudyPlanItemID", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "error UnMarkItemCompleted",
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission4: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentSubmission := &entities.StudentSubmission{
					StudyPlanItemID: database.Text("study-plan-item-id"),
				}
				submissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentSubmission, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				submissionRepo.On("DeleteByStudyPlanItemIDs", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(nil)
				latestSubmissionRepo.On("DeleteByStudyPlanItemID", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UnMarkItemCompleted", mock.Anything, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentSubmission := &entities.StudentSubmission{
					StudyPlanItemID: database.Text("study-plan-item-id"),
				}
				submissionRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentSubmission, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				submissionRepo.On("DeleteByStudyPlanItemIDs", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(nil)
				latestSubmissionRepo.On("DeleteByStudyPlanItemID", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UnMarkItemCompleted", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.DeleteStudentSubmission(ctx, testCase.req.(*pb.DeleteStudentSubmissionRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

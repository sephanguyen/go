package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEntryExitModifierService_DeleteEntryExit(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockStudentEntryExitRecordsRepo := new(mock_repositories.MockStudentEntryExitRecordsRepo)
	mockStudentRepo := new(mock_repositories.MockStudentRepo)

	mockValidStudent := &entities.Student{
		ID: database.Text("test"),
	}

	s := &EntryExitModifierService{
		DB:                          mockDB,
		StudentEntryExitRecordsRepo: mockStudentEntryExitRecordsRepo,
		StudentRepo:                 mockStudentRepo,
	}

	testcases := []TestCase{
		{
			name:        "happy case for deleting entry and exit record",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &eepb.DeleteEntryExitRequest{
				EntryexitId: 1,
				StudentId:   uuid.New().String(),
			},
			expectedResp: &eepb.DeleteEntryExitResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockValidStudent, nil)
				mockStudentEntryExitRecordsRepo.On("SoftDeleteByID", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "delete entry and exit record failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, "tx is closed"),
			req: &eepb.DeleteEntryExitRequest{
				EntryexitId: 1,
				StudentId:   uuid.New().String(),
			},
			setup: func(ctx context.Context) {
				mockStudentRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockValidStudent, nil)
				mockStudentEntryExitRecordsRepo.On("SoftDeleteByID", ctx, mockDB, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "delete entry and exit failed due to invalid entry and exit id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "invalid entry exit id"),
			req: &eepb.DeleteEntryExitRequest{
				EntryexitId: 0,
				StudentId:   uuid.New().String(),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "delete entry and exit failed due to invalid student id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "student id cannot be empty"),
			req: &eepb.DeleteEntryExitRequest{
				EntryexitId: 1,
				StudentId:   "",
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.DeleteEntryExit(testCase.ctx, testCase.req.(*eepb.DeleteEntryExitRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}

			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.(*eepb.DeleteEntryExitResponse).Successful, resp.Successful)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockStudentEntryExitRecordsRepo, mockStudentRepo)
		})
	}
}

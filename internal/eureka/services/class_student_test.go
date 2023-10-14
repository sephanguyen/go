package services

import (
	"context"
	"testing"
	"time"

	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassService_JoinClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockClassStudentRepo := new(mock_repositories.MockClassStudentRepo)
	type request struct {
		classId int32
		userId  string
	}

	s := &ClassStudentService{
		DB:               db,
		ClassStudentRepo: mockClassStudentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &request{
				classId: 12,
				userId:  "user-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockClassStudentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*request)
			err := s.upsertClassStudent(testCase.ctx, &pb.EvtClassRoom_JoinClass{
				ClassId: req.classId,
				UserId:  req.userId,
			})
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}

}

func TestClassService_LeaveClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockClassStudentRepo := new(mock_repositories.MockClassStudentRepo)
	type request struct {
		classId int32
		userId  string
	}

	s := &ClassStudentService{
		DB:               db,
		ClassStudentRepo: mockClassStudentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &request{
				classId: 12,
				userId:  "user-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockClassStudentRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*request)
			err := s.softDeleteClassMember(testCase.ctx, &pb.EvtClassRoom_LeaveClass{
				ClassId: req.classId,
				UserIds:  []string{req.userId},
			})
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}

}

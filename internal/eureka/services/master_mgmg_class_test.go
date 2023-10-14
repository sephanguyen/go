package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMasterMgmtClassService_CreateCourseClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	t.Run("[ActionKind UPSERT] should create CourseClass success", func(t *testing.T) {
		// Arrange
		courseClassRepo := &mock_repositories.MockCourseClassRepo{}
		c := &MasterMgmtClassService{
			DB:              db,
			CourseClassRepo: courseClassRepo,
		}
		courseIDs := []string{"course-id"}
		courseClassRepo.On("BulkUpsert", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseClass)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].ClassID.String, "class-id")
			}).
			Return(nil)

		// Action
		err := c.upsertCourseClass(ctx, &mpb.EvtClass_CreateClass{
			ClassId:  "class-id",
			CourseId: "course-id",
		})

		// Assert
		assert.Nil(t, err)
	})
}

func TestMasterMgmtClassService_JoinClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockMasterMgmtClassStudentRepo := new(mock_repositories.MockMasterMgmtClassStudentRepo)
	type request struct {
		classId string
		userId  string
	}

	s := &MasterMgmtClassService{
		DB:                         db,
		MasterMgmtClassStudentRepo: mockMasterMgmtClassStudentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &request{
				classId: "class-id",
				userId:  "user-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockMasterMgmtClassStudentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*request)
			err := s.upsertClassStudent(testCase.ctx, &mpb.EvtClass_JoinClass{
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

func TestMasterMgmtClassService_LeaveClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockMasterMgmtClassStudentRepo := new(mock_repositories.MockMasterMgmtClassStudentRepo)
	type request struct {
		classId string
		userId  string
	}

	s := &MasterMgmtClassService{
		DB:                         db,
		MasterMgmtClassStudentRepo: mockMasterMgmtClassStudentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &request{
				classId: "class-id",
				userId:  "user-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockMasterMgmtClassStudentRepo.On("SoftDelete", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*request)
			err := s.softDeleteClassMember(testCase.ctx, &mpb.EvtClass_LeaveClass{
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

func TestMasterMgmtClassService_DeleteClass(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := mock_database.Ext{}
	courseClassRepo := &mock_repositories.MockCourseClassRepo{}
	s := &MasterMgmtClassService{
		DB:                         &db,
		MasterMgmtClassStudentRepo: new(mock_repositories.MockMasterMgmtClassStudentRepo),
		CourseClassRepo:            courseClassRepo,
	}

	classID := "Class-ID"

	testCases := []TestCase{
		{
			name: "Happy case delete single class",
			req: &mpb.EvtClass_DeleteClass{
				ClassId: classID,
			},
			setup: func(ctx context.Context) {
				courseClassRepo.On("DeleteClass", ctx, s.DB, classID).Once().Return(nil)
			},
		},
		{
			name: "Error when classID is empty",
			req: &mpb.EvtClass_DeleteClass{
				ClassId: "",
			},
			setup: func(ctx context.Context) {
				courseClassRepo.AssertNotCalled(t, "DeleteClass")
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("cannot empty class_id").Error()),
		},
		{
			name: "Error when DeleteClass err",
			req: &mpb.EvtClass_DeleteClass{
				ClassId: classID,
			},
			setup: func(ctx context.Context) {
				courseClassRepo.On("DeleteClass", ctx, s.DB, classID).Once().Return(puddle.ErrClosedPool)
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.CourseClassRepo.DeleteClass: %w", puddle.ErrClosedPool).Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.req.(*mpb.EvtClass_DeleteClass)
			testCase.setup(ctx)

			err := s.deleteClass(ctx, input)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, testCase.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}

}

func TestMasterMgmtClassService_HandleMasterMgmtClassEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	mockMasterMgmtClassStudentRepo := new(mock_repositories.MockMasterMgmtClassStudentRepo)
	courseClassRepo := &mock_repositories.MockCourseClassRepo{}
	s := &MasterMgmtClassService{
		DB:                         db,
		MasterMgmtClassStudentRepo: mockMasterMgmtClassStudentRepo,
		CourseClassRepo:            courseClassRepo,
	}

	classID := "Class-ID"

	testCases := []TestCase{
		{
			name:        "create class",
			ctx:         ctx,
			req:         &mpb.EvtClass{Message: &mpb.EvtClass_CreateClass_{}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseClassRepo.On("BulkUpsert", mock.Anything, db, mock.AnythingOfType("[]*entities.CourseClass")).Return(nil)
			},
		},
		{
			name: "join class",
			ctx:  ctx,
			req: &mpb.EvtClass{Message: &mpb.EvtClass_JoinClass_{
				JoinClass: &mpb.EvtClass_JoinClass{
					ClassId: "class-id",
					UserId:  "user-id",
				},
			}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockMasterMgmtClassStudentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "leave class",
			ctx:         ctx,
			req:         &mpb.EvtClass{Message: &mpb.EvtClass_LeaveClass_{}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "delete class success",
			ctx:  ctx,
			req: &mpb.EvtClass{Message: &mpb.EvtClass_DeleteClass_{
				DeleteClass: &mpb.EvtClass_DeleteClass{
					ClassId: classID,
				},
			}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseClassRepo.On("DeleteClass", ctx, s.DB, classID).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.EvtClass)
			err := s.HandleMasterMgmtClassEvent(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

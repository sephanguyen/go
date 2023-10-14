package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentReader_RetrieveStudentAssociatedToParentAccount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	student := generateEnUser()
	studentRepository := &mock_repositories.MockStudentRepo{}
	s := &UserReaderService{
		StudentRepo: studentRepository,
		DB:          db,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req:  &pb.RetrieveStudentAssociatedToParentAccountRequest{},
			setup: func(ctx context.Context) {
				studentRepository.On("GetStudentsByParentID", ctx, db, database.Text("id")).
					Return([]*entity.LegacyUser{student}, nil).
					Once()
			},
			expectedErr: nil,
		},
		{
			name: "retrieve empty student",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req:  &pb.RetrieveStudentAssociatedToParentAccountRequest{},
			setup: func(ctx context.Context) {
				studentRepository.On("GetStudentsByParentID", ctx, db, database.Text("id")).
					Return([]*entity.LegacyUser{}, nil).
					Once()
			},
			expectedErr: nil,
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.RetrieveStudentAssociatedToParentAccountRequest{},
			expectedErr: fmt.Errorf("failed to get students by parentID: %v", fmt.Errorf("error query")),
			setup: func(ctx context.Context) {
				studentRepository.On("GetStudentsByParentID", ctx, db, database.Text("id")).
					Return(nil, fmt.Errorf("error query")).
					Once()
			},
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.RetrieveStudentAssociatedToParentAccountRequest{},
			expectedErr: fmt.Errorf("failed to get students by parentID: %v", fmt.Errorf("error query")),
			setup: func(ctx context.Context) {
				studentRepository.On("GetStudentsByParentID", ctx, db, database.Text("id")).
					Return(nil, fmt.Errorf("error query")).
					Once()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.RetrieveStudentAssociatedToParentAccountRequest)
			_, err := s.RetrieveStudentAssociatedToParentAccount(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepository)
		})
	}
}

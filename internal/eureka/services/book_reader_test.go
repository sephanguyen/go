package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBookReader_RetrieveLOs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = interceptors.ContextWithUserGroup(ctx, usermgmt_entities.UserGroupStudent)
	studentStudyPlanRepo := new(mock_repositories.MockStudentStudyPlanRepo)
	bookRepo := new(mock_repositories.MockBookRepo)
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.ListBooksRequest{
				Filter: &cpb.CommonFilter{
					Ids: []string{"book-1", "book-2"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentStudyPlanRepo.On("GetBookIDsBelongsToStudentStudyPlan", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"book-1", "book-2"}, nil)
				bookRepo.On("ListBooks", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.Book{
					{
						ID: database.Text("book-1"),
					},
					{
						ID: database.Text("book-2"),
					},
				}, nil)
			},
		},
		{
			name: "repo err case",
			ctx:  ctx,
			req: &epb.ListBooksRequest{
				Filter: &cpb.CommonFilter{
					Ids: []string{"book-1", "book-2"},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(pgx.ErrNoRows, "crs.StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan").Error()),
			setup: func(ctx context.Context) {
				studentStudyPlanRepo.On("GetBookIDsBelongsToStudentStudyPlan", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	s := &BookReaderService{
		StudentStudyPlanRepo: studentStudyPlanRepo,
		BookRepo:             bookRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.ListBooksRequest)
			_, err := s.ListBooks(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

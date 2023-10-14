package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpsertBooks_Error(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	bookRepo := &mock_repositories.MockBookRepo{}

	courseService := &BookModifierService{
		DBTrace:  &database.DBTrace{DB: mockDB},
		BookRepo: bookRepo,
	}

	testCases := map[string]TestCase{
		"missing name": {
			req: &pb.UpsertBooksRequest{
				Books: []*pb.UpsertBooksRequest_Book{{
					BookId: "id",
					Name:   "",
				}},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateBook: name cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		"happy case": {
			req: &pb.UpsertBooksRequest{
				Books: []*pb.UpsertBooksRequest_Book{{
					BookId: "id",
					Name:   "name",
				}},
			},
			setup: func(ctx context.Context) {
				bookRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)

			_, err := courseService.UpsertBooks(ctx, testCase.req.(*pb.UpsertBooksRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestNewBookModifierService(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		mockDB := &mock_database.Ext{}
		s := NewBookModifierService(mockDB).(*BookModifierService)
		assert.IsType(t, new(repositories.BookRepo), s.BookRepo)
	})
}

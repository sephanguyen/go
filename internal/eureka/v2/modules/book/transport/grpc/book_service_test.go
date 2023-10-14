package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/book/usecase/repo"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestBookService_UpsertBooks(t *testing.T) {
	t.Parallel()

	bookUsecase := &mock_usecase.MockBookUsecase{}

	bookSvc := &BookService{
		UpsertBooksUsecase: bookUsecase,
		BookContentUsecase: bookUsecase,
	}

	testCases := map[string]TestCase{
		"missing name": {
			req: &pb.UpsertBooksRequest{
				Books: []*pb.UpsertBooksRequest_Book{{
					BookId: "id",
					Name:   "",
				}},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.NewConversionError("name cannot be empty", nil).Error()),
		},
		"happy case": {
			req: &pb.UpsertBooksRequest{
				Books: []*pb.UpsertBooksRequest_Book{{
					BookId: "id",
					Name:   "name",
				}},
			},
			setup: func(ctx context.Context) {
				bookUsecase.On("UpsertBooks", ctx, mock.Anything).
					Once().
					Run(func(args mock.Arguments) {
						books := args[1].([]domain.Book)
						assert.Equal(t, "id", books[0].ID)
						assert.Equal(t, "name", books[0].Name)
					}).
					Return(nil)
			},
			expectedResp: &pb.UpsertBooksResponse{
				BookIds: []string{"id"},
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			if testCase.setup != nil {
				testCase.setup(ctx)
			}

			resp, err := bookSvc.UpsertBooks(ctx, testCase.req.(*pb.UpsertBooksRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.UpsertBooksResponse), resp)
			}
		})
	}
}

func TestBookService_GetBookContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("book id is not found then return grpc not found error", func(t *testing.T) {
		t.Parallel()
		// arrange
		bookUsecase := &mock_usecase.MockBookUsecase{}
		bookSvc := &BookService{
			UpsertBooksUsecase: bookUsecase,
			BookContentUsecase: bookUsecase,
		}
		req := &pb.GetBookContentRequest{
			BookId: "book-id",
		}
		usecaseErr := errors.NewEntityNotFoundError("BookUsecase.GetPublishedBookContent", nil)
		expectedErr := status.Error(codes.NotFound, usecaseErr.Error())
		bookUsecase.On("GetPublishedBookContent", ctx, "book-id").
			Once().
			Return(domain.Book{}, usecaseErr)

		// act
		resp, err := bookSvc.GetBookContent(ctx, req)

		// assert
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
		mock.AssertExpectationsForObjects(t, bookUsecase)
	})

	t.Run("return book content successfully", func(t *testing.T) {
		t.Parallel()
		// arrange
		bookUsecase := &mock_usecase.MockBookUsecase{}
		bookSvc := &BookService{
			UpsertBooksUsecase: bookUsecase,
			BookContentUsecase: bookUsecase,
		}
		book := domain.Book{
			ID:   "book-id",
			Name: "some name",
			Chapters: []domain.Chapter{
				{
					ID:   "chapter id",
					Name: "chapter name",
					Topics: []domain.Topic{
						{
							ID:   "topic id",
							Name: "topic name",
							LearningMaterials: []domain.LearningMaterial{
								{
									ID:        "LM ID",
									Name:      "LM ID",
									Published: true,
								},
								{
									ID:        "LM ID 2",
									Name:      "LM ID 2",
									Published: false,
								},
							},
						},
					},
				},
			},
		}
		book.RemoveUnpublishedContent()
		req := &pb.GetBookContentRequest{
			BookId: book.ID,
		}
		expectedResp := pb.GetBookContentResponse{
			Id:       book.ID,
			Name:     book.Name,
			Chapters: transformChapterToContentPb(book.Chapters),
		}

		bookUsecase.On("GetPublishedBookContent", ctx, "book-id").
			Once().
			Return(book, nil)

		// act
		resp, err := bookSvc.GetBookContent(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResp.Id, resp.Id)
		assert.Equal(t, expectedResp.Name, resp.Name)
		assert.Equal(t, expectedResp.Chapters, resp.Chapters)
		mock.AssertExpectationsForObjects(t, bookUsecase)
	})
}

func TestBookService_GetBookHierarchyFlattenByLearningMaterialID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	bookUsecase := &mock_usecase.MockBookUsecase{}
	bookSvc := &BookService{
		BookHierarchyUsecase: bookUsecase,
	}
	learningMaterialID := "learningMaterial_9"

	t.Run("happy case", func(t *testing.T) {
		// arrange
		req := &pb.GetBookHierarchyFlattenByLearningMaterialIDRequest{
			LearningMaterialId: learningMaterialID,
		}

		bookHierarchyFlatten := domain.BookHierarchyFlatten{
			BookID:             "BookID",
			ChapterID:          "ChapterID",
			TopicID:            "TopicID",
			LearningMaterialID: "LearningMaterialID",
		}

		// act
		bookUsecase.On("GetBookHierarchyFlattenByLearningMaterialID", ctx, learningMaterialID).Once().Return(bookHierarchyFlatten, nil)

		res, err := bookSvc.GetBookHierarchyFlattenByLearningMaterialID(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, transformBookHierarchyFlattenToGetBookHierarchyByLearningMaterialIDResponse(bookHierarchyFlatten), res)
	})

	t.Run("error on usecase.GetBookHierarchyFlattenByLearningMaterialID", func(t *testing.T) {
		// arrange
		req := &pb.GetBookHierarchyFlattenByLearningMaterialIDRequest{
			LearningMaterialId: learningMaterialID,
		}

		bookHierarchyFlatten := domain.BookHierarchyFlatten{}
		usecaseErr := errors.NewEntityNotFoundError("GetBookHierarchyFlattenByLearningMaterialID error", nil)

		// act
		bookUsecase.On("GetBookHierarchyFlattenByLearningMaterialID", ctx, learningMaterialID).Once().Return(bookHierarchyFlatten, usecaseErr)

		res, err := bookSvc.GetBookHierarchyFlattenByLearningMaterialID(ctx, req)

		// assert
		assert.Nil(t, res)
		assert.Equal(t, errors.NewGrpcError(usecaseErr, transport.GrpcErrorMap), err)
	})

	t.Run("transform data to response correctly", func(t *testing.T) {
		// arrange
		bookHierarchyFlatten := domain.BookHierarchyFlatten{
			BookID:             "BookID_1",
			ChapterID:          "ChapterID_2",
			TopicID:            "TopicID_9",
			LearningMaterialID: "LearningMaterialID_10",
		}

		expected := &pb.GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten{
			BookId:             bookHierarchyFlatten.BookID,
			ChapterId:          bookHierarchyFlatten.ChapterID,
			TopicId:            bookHierarchyFlatten.TopicID,
			LearningMaterialId: bookHierarchyFlatten.LearningMaterialID,
		}

		// act
		actual := transformBookHierarchyFlattenToGetBookHierarchyByLearningMaterialIDResponse(bookHierarchyFlatten)

		// assert
		assert.Equal(t, expected, actual.BookHierarchyFlatten)
	})
}

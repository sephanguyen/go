package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

type BookService struct {
	UpsertBooksUsecase   usecase.BookUpserter
	BookContentUsecase   usecase.BookContentGetter
	BookHierarchyUsecase usecase.BookHierarchyGetter
}

func NewBookService(
	bookUsecase *usecase.BookUsecase,
) *BookService {
	return &BookService{
		UpsertBooksUsecase:   bookUsecase,
		BookContentUsecase:   bookUsecase,
		BookHierarchyUsecase: bookUsecase,
	}
}

func (b *BookService) UpsertBooks(ctx context.Context, req *pb.UpsertBooksRequest) (*pb.UpsertBooksResponse, error) {
	books := make([]domain.Book, len(req.Books))
	bookIDs := make([]string, len(req.Books))

	for i, v := range req.Books {
		book := transformBookFromPb(v)
		if err := validateBook(book); err != nil {
			return &pb.UpsertBooksResponse{}, errors.NewGrpcError(err, transport.GrpcErrorMap)
		}
		books[i] = book
		bookIDs[i] = book.ID
	}

	if err := b.UpsertBooksUsecase.UpsertBooks(ctx, books); err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}
	return &pb.UpsertBooksResponse{
		BookIds: bookIDs,
	}, nil
}

func transformBookFromPb(req *pb.UpsertBooksRequest_Book) domain.Book {
	if req.BookId == "" {
		req.BookId = idutil.ULIDNow()
	}

	return domain.NewBook(req.BookId, req.Name)
}

func validateBook(book domain.Book) error {
	if book.Name == "" {
		return errors.NewConversionError("name cannot be empty", nil)
	}
	return nil
}

func (b *BookService) GetBookContent(ctx context.Context, req *pb.GetBookContentRequest) (*pb.GetBookContentResponse, error) {
	book, err := b.BookContentUsecase.GetPublishedBookContent(ctx, req.GetBookId())
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pb.GetBookContentResponse{
		Id:       book.ID,
		Name:     book.Name,
		Chapters: transformChapterToContentPb(book.Chapters),
	}, nil
}

func transformChapterToContentPb(chapters []domain.Chapter) []*pb.GetBookContentResponse_Chapter {
	var convertLMs = func(materials []domain.LearningMaterial) []*pb.GetBookContentResponse_LearningMaterial {
		return sliceutils.Map(materials, func(m domain.LearningMaterial) *pb.GetBookContentResponse_LearningMaterial {
			return &pb.GetBookContentResponse_LearningMaterial{
				Id:           m.ID,
				DisplayOrder: int32(m.DisplayOrder),
				Name:         m.Name,
				Type:         m.Type.GetProtobufType(),
			}
		})
	}
	var convertTopics = func(topics []domain.Topic) []*pb.GetBookContentResponse_Topic {
		return sliceutils.Map(topics, func(t domain.Topic) *pb.GetBookContentResponse_Topic {
			return &pb.GetBookContentResponse_Topic{
				Id:                t.ID,
				Name:              t.Name,
				DisplayOrder:      int32(t.DisplayOrder),
				IconUrl:           t.IconURL,
				LearningMaterials: convertLMs(t.LearningMaterials),
			}
		})
	}

	return sliceutils.Map(chapters, func(c domain.Chapter) *pb.GetBookContentResponse_Chapter {
		return &pb.GetBookContentResponse_Chapter{
			Id:           c.ID,
			Name:         c.Name,
			DisplayOrder: int32(c.DisplayOrder),
			Topics:       convertTopics(c.Topics),
		}
	})
}

func (b *BookService) GetBookHierarchyFlattenByLearningMaterialID(ctx context.Context, req *pb.GetBookHierarchyFlattenByLearningMaterialIDRequest) (*pb.GetBookHierarchyFlattenByLearningMaterialIDResponse, error) {
	hierarchyFlatten, err := b.BookHierarchyUsecase.GetBookHierarchyFlattenByLearningMaterialID(ctx, req.GetLearningMaterialId())
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return transformBookHierarchyFlattenToGetBookHierarchyByLearningMaterialIDResponse(hierarchyFlatten), nil
}

func transformBookHierarchyFlattenToGetBookHierarchyByLearningMaterialIDResponse(hierarchyFlatten domain.BookHierarchyFlatten) *pb.GetBookHierarchyFlattenByLearningMaterialIDResponse {
	return &pb.GetBookHierarchyFlattenByLearningMaterialIDResponse{
		BookHierarchyFlatten: &pb.GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten{
			BookId:             hierarchyFlatten.BookID,
			ChapterId:          hierarchyFlatten.ChapterID,
			TopicId:            hierarchyFlatten.TopicID,
			LearningMaterialId: hierarchyFlatten.LearningMaterialID,
		},
	}
}

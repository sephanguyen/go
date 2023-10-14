package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookModifierService struct {
	DBTrace database.Ext

	BookRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
	}
}

func NewBookModifierService(db database.Ext) pb.BookModifierServiceServer {
	return &BookModifierService{
		DBTrace:  db,
		BookRepo: new(repositories.BookRepo),
	}
}

func validateBook(c *pb.UpsertBooksRequest_Book) error {
	if c.Name == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}

func toBookEntity(req *pb.UpsertBooksRequest_Book) (*entities.Book, error) {
	if req.BookId == "" {
		req.BookId = idutil.ULIDNow()
	}

	r := &entities.Book{}
	database.AllNullEntity(r)
	if err := multierr.Combine(
		r.ID.Set(req.BookId),
		r.Name.Set(req.Name),
		r.Country.Set(cpb.Country_COUNTRY_NONE),
		r.Subject.Set(cpb.Subject_SUBJECT_NONE),
		r.Grade.Set(0),
		r.CurrentChapterDisplayOrder.Set(0), // set initial value = 0
		r.BookType.Set(cpb.BookType_BOOK_TYPE_GENERAL.String()),
		r.IsV2.Set(false), // Set is v2 is false because this endpoint just apply to book v1
	); err != nil {
		return nil, err
	}

	return r, nil
}

func toEnBookChapter(chapterID []string, bookID string) ([]*entities.BookChapter, error) {
	result := []*entities.BookChapter{}
	for _, chapterID := range chapterID {
		if chapterID == "" {
			continue
		}
		r := &entities.BookChapter{}

		database.AllNullEntity(r)
		if err := multierr.Combine(
			r.BookID.Set(bookID),
			r.ChapterID.Set(chapterID),
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

func (s *BookModifierService) UpsertBooks(ctx context.Context, req *pb.UpsertBooksRequest) (*pb.UpsertBooksResponse, error) {
	books := []*entities.Book{}
	bookIDs := []string{}

	for _, v := range req.Books {
		if err := validateBook(v); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateBook: %w", err).Error())
		}

		book, err := toBookEntity(v)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("toBookEntity: %w", err).Error())
		}

		books = append(books, book)
		bookIDs = append(bookIDs, book.ID.String)
	}

	if err := s.BookRepo.Upsert(ctx, s.DBTrace, books); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("BookRepo.Upsert: %w", err).Error())
	}

	return &pb.UpsertBooksResponse{
		BookIds: bookIDs,
	}, nil
}

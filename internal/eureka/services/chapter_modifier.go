package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChapterModifierService struct {
	DBTrace database.Ext

	BookRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Book, error)
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
	}

	ChapterRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs []string) (int, error)
	}

	BookChapterRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) error
	}
}

func NewChapterModifierService(db database.Ext) *ChapterModifierService {
	return &ChapterModifierService{
		DBTrace:         db,
		BookRepo:        new(repositories.BookRepo),
		ChapterRepo:     new(repositories.ChapterRepo),
		BookChapterRepo: new(repositories.BookChapterRepo),
	}
}

func validateChapter(src *cpb.Chapter) error {
	if src.Info == nil {
		return status.Error(codes.InvalidArgument, "chapter info cannot be empty")
	}

	if src.Info.Name == "" {
		return status.Error(codes.InvalidArgument, "chapter name cannot be empty")
	}

	if src.Info.SchoolId == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("missing school id of chapter %s", src.Info.Name))
	}

	if src.Info.DisplayOrder < 0 {
		return status.Error(codes.InvalidArgument, "display_order cannot be less than 0")
	}

	return nil
}

func convertChapterProto2Entity(src *cpb.Chapter) (*entities.Chapter, error) {
	if src.Info.Id == "" {
		src.Info.Id = idutil.ULIDNow()
	}
	r := &entities.Chapter{}
	database.AllNullEntity(r)
	if err := multierr.Combine(
		r.ID.Set(src.Info.Id),
		r.Name.Set(src.Info.Name),
		r.Country.Set(src.Info.Country.String()),
		r.DisplayOrder.Set(src.Info.DisplayOrder),
		r.SchoolID.Set(src.Info.SchoolId),
		r.Grade.Set(src.Info.Grade),
		r.Subject.Set(src.Info.Subject.String()),
		r.DeletedAt.Set(nil),
		r.CurrentTopicDisplayOrder.Set(0),
		r.BookID.Set(src.BookId),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to set entities data of chapter: %v", err))
	}

	return r, nil
}

func isAutomaticDisplayOrder(req *epb.UpsertChaptersRequest) bool {
	if req.BookId == "" {
		return false
	}
	for _, c := range req.GetChapters() {
		if c.Info.DisplayOrder != 0 {
			return false
		}
	}
	return true
}

func addDisplayOrderAccording(src *entities.Chapter, dOrder int32) error {
	return src.DisplayOrder.Set(dOrder)
}

// UpsertChapters have two cases: manual set display order and automatically set display_order
// step 1: check the request send display_order or not. if yes -> still use old flow, else -> next step 2
// step 2: get the current display_order -> update for each chapter and upsert
func (s *ChapterModifierService) UpsertChapters(ctx context.Context, req *epb.UpsertChaptersRequest) (*epb.UpsertChaptersResponse, error) {
	chapters := []*entities.Chapter{}
	chapterIDs := []string{}

	for _, v := range req.Chapters {
		if err := validateChapter(v); err != nil {
			return nil, err
		}

		chapter, err := convertChapterProto2Entity(v)
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, chapter)
		chapterIDs = append(chapterIDs, chapter.ID.String)
	}

	if !isAutomaticDisplayOrder(req) {
		if err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
			if req.BookId == "" { // if not have book id -> only upsert
				if err := s.ChapterRepo.Upsert(ctx, tx, chapters); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.ChapterRepo.Upsert: %w", err).Error())
				}
			} else {
				// if yes -> lock book -> upsert -> upsert book_chapter -> update current
				if _, err := s.BookRepo.FindByID(ctx, tx, database.Text(req.BookId), repositories.WithUpdateLock()); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("BookRepo.FindByID: %w", err).Error())
				}

				chapterMaps, err := s.ChapterRepo.FindByIDs(ctx, tx, chapterIDs)
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("ChapterRepo.FindByIDs: %w", err).Error())
				}

				totalChapterInserted := len(chapters) - len(chapterMaps)

				if err = s.ChapterRepo.Upsert(ctx, tx, chapters); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.ChapterRepo.Upsert: %w", err).Error())
				}

				books, err := s.BookRepo.FindByIDs(ctx, tx, []string{req.BookId})
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.BookRepo.FindByIDs: %w", err).Error())
				}

				if _, ok := books[req.BookId]; !ok {
					return status.Error(codes.NotFound, "book not found")
				}

				bookChapters, err := toEnBookChapter(chapterIDs, req.BookId)
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("toEnBookChapter: %w", err).Error())
				}

				if err = s.BookChapterRepo.Upsert(ctx, tx, bookChapters); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.BookChapterRepo.Upsert: %w", err).Error())
				}

				if err = s.BookRepo.UpdateCurrentChapterDisplayOrder(ctx, tx, database.Int4(int32(totalChapterInserted)), database.Text(req.GetBookId())); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("BookRepo.UpdateCurrentChapterDisplayOrder: %w", err).Error())
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		if err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
			book, err := s.BookRepo.FindByID(ctx, tx, database.Text(req.GetBookId()), repositories.WithUpdateLock())
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("BookRepo.FindByID: %w", err).Error())
			}
			currenDisplayOrder := book.CurrentChapterDisplayOrder.Int
			var totalGeneratedChapterDisplayOrder int32 = 0
			// because when we upsert, only automatically display_order which not existed on database
			chapterMaps, err := s.ChapterRepo.FindByIDs(ctx, tx, chapterIDs)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("ChapterRepo.FindByIDs: %w", err).Error())
			}
			for _, c := range chapters {
				// only automatically display_order which not existed on database
				if _, ok := chapterMaps[c.ID.String]; !ok {
					totalGeneratedChapterDisplayOrder++
					if err := addDisplayOrderAccording(c, currenDisplayOrder+totalGeneratedChapterDisplayOrder); err != nil {
						return status.Error(codes.Internal, fmt.Errorf("addDisplayOrderAccording: %w", err).Error())
					}
				}
			}
			// upsert chapter
			if err = s.ChapterRepo.UpsertWithoutDisplayOrderWhenUpdate(ctx, tx, chapters); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.ChapterRepo.Upsert: %w", err).Error())
			}

			bookChapters, err := toEnBookChapter(chapterIDs, req.BookId)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("toEnBookChapter: %w", err).Error())
			}

			if err = s.BookChapterRepo.Upsert(ctx, tx, bookChapters); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.BookChapterRepo.Upsert: %w", err).Error())
			}
			// update to current
			if err = s.BookRepo.UpdateCurrentChapterDisplayOrder(ctx, tx, database.Int4(totalGeneratedChapterDisplayOrder), database.Text(req.GetBookId())); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("BookRepo.UpdateCurrentChapterDisplayOrder: %w", err).Error())
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &epb.UpsertChaptersResponse{
		Successful: true,
		ChapterIds: chapterIDs,
	}, nil
}

// DeleteChapters have some steps:
// Soft delete chapters
// Soft delete book-chapters
func (s *ChapterModifierService) DeleteChapters(ctx context.Context, req *epb.DeleteChaptersRequest) (*epb.DeleteChaptersResponse, error) {
	chapterMap, err := s.ChapterRepo.FindByIDs(ctx, s.DBTrace, req.ChapterIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to find chapter by ids: %w", err).Error())
	}

	for _, chapterID := range req.ChapterIds {
		if _, ok := chapterMap[chapterID]; !ok {
			return nil, status.Errorf(codes.InvalidArgument, "chapter %v does not exists", chapterID)
		}
	}
	if err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.ChapterRepo.SoftDelete(ctx, tx, req.ChapterIds); err != nil {
			return fmt.Errorf("unable to delete chapters: %w", err)
		}

		if err := s.BookChapterRepo.SoftDeleteByChapterIDs(ctx, tx, req.ChapterIds); err != nil {
			return fmt.Errorf("unable to delete books chapters: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.DeleteChaptersResponse{Successful: true}, nil
}

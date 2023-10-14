package book

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

func (s *Suite) aUserInsertABookToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	book := &entities.Book{}
	database.AllNullEntity(book)
	stepState.BookID = idutil.ULIDNow()
	now := timeutil.Now()
	if err := multierr.Combine(book.ID.Set(stepState.BookID),
		book.BookType.Set(cpb.BookType_BOOK_TYPE_GENERAL),
		book.Name.Set(fmt.Sprintf("book_name+%s", stepState.BookID)),
		book.SchoolID.Set(stepState.DefaultSchoolID),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup a book: %w", err)
	}
	bookRepo := repositories.BookRepo{}
	if err := bookRepo.Upsert(ctx, s.DB, []*entities.Book{book}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a book: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetABookByID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookRepo := repositories.BookRepo{}
	if book, err := bookRepo.FindByID(ctx, s.DB, database.Text(stepState.BookID)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to FindByID: %w", err)
	} else {
		stepState.Book = book
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnTheBookCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if stepState.Book.ID.String != stepState.BookID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book: want %s, actual: %s", stepState.BookID, stepState.Book.ID.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

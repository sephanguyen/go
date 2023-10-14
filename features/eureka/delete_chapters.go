package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) someMissingChapterIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < rand.Intn(5)+3; i++ {
		stepState.ChapterIDs = append(stepState.ChapterIDs, idutil.ULIDNow())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteSomeChapters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = epb.NewChapterModifierServiceClient(s.Conn).DeleteChapters(ctx, &epb.DeleteChaptersRequest{
		ChapterIds: stepState.ChapterIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustDeleteTheChaptersCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ce := entities.Chapter{}
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE chapter_id = ANY($1::TEXT[]) AND deleted_at IS NOT NULL`, ce.TableName())
	var softDeletedChaptersCount int
	if err := s.DB.QueryRow(ctx, query, database.TextArray(stepState.ChapterIDs)).Scan(&softDeletedChaptersCount); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query soft deleted chapters count: %w", err)
	}
	if softDeletedChaptersCount != len(stepState.ChapterIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to soft deleted chapters expected: %v, got: %v", len(stepState.ChapterIDs), softDeletedChaptersCount)
	}

	bce := entities.BookChapter{}
	var softDeletedBookChaptersCount int
	query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE book_id = $1 AND chapter_id = ANY($2::TEXT[]) AND deleted_at IS NOT NULL`, bce.TableName())
	if err := s.DB.QueryRow(ctx, query, database.Text(stepState.BookID), database.TextArray(stepState.ChapterIDs)).Scan(&softDeletedBookChaptersCount); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query soft deleted book_chapters count: %w", err)
	}
	// Because of chapter relate to only 1 book, so simple check the number of created chapters and number of  deleted book_chapters.
	// Need to update this if we change the relation of book and chapter in the future.
	if softDeletedBookChaptersCount != len(stepState.ChapterIDs) {
		fmt.Printf("BookID: '%s'", stepState.BookID)
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to soft deleted book_chapters expected: %v, got: %v", len(stepState.ChapterIDs), softDeletedBookChaptersCount)
	}

	return StepStateToContext(ctx, stepState), nil
}

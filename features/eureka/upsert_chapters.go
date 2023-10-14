package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/proto"
)

func (s *suite) generateChapters(ctx context.Context, numberOfChapter int, template *cpb.Chapter) []*cpb.Chapter {
	stepState := StepStateFromContext(ctx)
	if template == nil {
		template = &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Country:  cpb.Country_COUNTRY_VN,
				SchoolId: constants.ManabieSchool,
				Subject:  cpb.Subject_SUBJECT_BIOLOGY,
				Grade:    1,
				Name:     "Chapter",
			},
			BookId: stepState.BookID,
		}
	}
	chapters := make([]*cpb.Chapter, 0)
	for i := 0; i < numberOfChapter; i++ {
		chapters = append(chapters, proto.Clone(template).(*cpb.Chapter))
	}
	return chapters
}

func (s *suite) userCreateAValidChapter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	reqChapters := s.generateChapters(ctx, 1, nil)
	res, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: reqChapters,
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to create a chapter: %w", err)
	}
	if len(res.ChapterIds) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unexpected number of topics: want 1, got %d", len(res.ChapterIds))
	}
	stepState.ChapterID = res.ChapterIds[0]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreatesNewChapters(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfChapter := rand.Intn(10) + 5
	var reqChapters []*cpb.Chapter
	switch validity {
	case "valid":
		reqChapters = s.generateChapters(ctx, numberOfChapter, nil)
	case "invalid":
		reqChapters = s.generateChapters(ctx, numberOfChapter, &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Country:  cpb.Country_COUNTRY_VN,
				SchoolId: 0,
				Subject:  0,
				Name:     "",
				Grade:    0,
			},
			BookId: stepState.BookID,
		})
	}
	stepState.Response, stepState.ResponseErr = epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: reqChapters,
		BookId:   stepState.BookID,
	})
	if stepState.ResponseErr == nil {
		stepState.ChapterIDs = stepState.Response.(*epb.UpsertChaptersResponse).ChapterIds
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustStoresCorrectChapters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT chapter_id,book_id FROM chapters WHERE chapter_id = ANY($1)"
	rows, err := s.DB.Query(ctx, query, stepState.ChapterIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to retrieve chapters: %w", err)
	}
	var count int
	for rows.Next() {
		var chapterID, bookID pgtype.Text
		if err := rows.Scan(&chapterID, &bookID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if bookID.String != stepState.BookID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("book_id of chapter is wrong, expect %v but got %v", stepState.BookID, bookID.String)
		}
		count++
	}

	if count != len(stepState.ChapterIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Chapters are not stored correctly: expected %d, got %d", len(stepState.ChapterIDs), count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreChaptersExisted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfChapters := rand.Intn(10) + 5
	reqChapters := s.generateChapters(ctx, numberOfChapters, nil)
	resp, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: reqChapters,
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to seeding some chapters: %w", err)
	}
	stepState.ChapterIDs = resp.ChapterIds

	query := "SELECT chapter_id, updated_at FROM chapters WHERE chapter_id = ANY($1)"
	rows, err := s.DB.Query(ctx, query, database.TextArray(stepState.ChapterIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query chapter: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		chapter := new(entities.Chapter)
		if err := rows.Scan(&chapter.ID, &chapter.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to scan chapter: %w", err)
		}
		stepState.Chapters = append(stepState.Chapters, chapter)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdatesChapters(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	// Take first <randomly> chapters to update, 1 at least
	numberOfChapters := rand.Intn(len(stepState.ChapterIDs)-1) + 1
	stepState.UpdatedChapterIDs = append(stepState.UpdatedChapterIDs, stepState.ChapterIDs[0:numberOfChapters]...)
	var reqChapters []*cpb.Chapter
	// Update data by each case.
	switch validity {
	case "valid":
		reqChapters = s.generateChapters(ctx, len(stepState.UpdatedChapterIDs), nil)
	case "invalid":
		reqChapters = s.generateChapters(ctx, len(stepState.UpdatedChapterIDs), &cpb.Chapter{
			BookId: stepState.BookID,
			Info: &cpb.ContentBasicInfo{
				Country: cpb.Country_COUNTRY_VN,
				// Missing Name, SchoolId, Subject, Grade
			},
		})
	}
	for i, chapterID := range stepState.UpdatedChapterIDs {
		reqChapters[i].Info.Id = chapterID
	}

	stepState.Response, stepState.ResponseErr = epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: reqChapters,
		BookId:   stepState.BookID,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateTheChaptersCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existedChapterMap := make(map[string]*entities.Chapter)
	for _, chapter := range stepState.Chapters {
		existedChapterMap[chapter.ID.String] = chapter
	}

	query := "SELECT chapter_id, updated_at FROM chapters WHERE chapter_id = ANY($1)"
	rows, err := s.DB.Query(ctx, query, database.TextArray(stepState.UpdatedChapterIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		chapter := new(entities.Chapter)
		if err := rows.Scan(&chapter.ID, &chapter.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query chapter: %w", err)
		}
		existedChapter, ok := existedChapterMap[chapter.ID.String]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Chapter is missing chapter_id: %s", chapter.ID.String)
		}
		if chapter.UpdatedAt.Time.Equal(existedChapter.UpdatedAt.Time) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Chapter was not updated chapter_id: %s", chapter.ID.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

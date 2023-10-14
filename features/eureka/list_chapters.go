package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) someChaptersAreExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	totalChaptersInEachBook := 12
	chapters := make([]*cpb.Chapter, 0, totalChaptersInEachBook)
	bookId := idutil.ULIDNow()
	book := &epb.UpsertBooksRequest_Book{
		BookId: bookId,
		Name:   strconv.Itoa(rand.Int()),
	}

	stepState.Request = &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{book},
	}

	stepState.Response, stepState.ResponseErr = epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(contextWithToken(s, ctx), stepState.Request.(*epb.UpsertBooksRequest))
	for j := 1; j <= totalChaptersInEachBook; j++ {
		chapter := &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Id:      idutil.ULIDNow(),
				Name:    "book-chapter-name-" + bookId,
				Country: cpb.Country_COUNTRY_NONE,
				Subject: cpb.Subject_SUBJECT_NONE,
				// Grade: "",
				DisplayOrder: int32(j),
				SchoolId:     constants.ManabieSchool,
			},
		}
		chapters = append(chapters, chapter)
	}
	_, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &epb.UpsertChaptersRequest{
		Chapters: chapters,
		BookId:   bookId,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapters: %w", err)
	}

	stepState.Request = chapters
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentListChaptersByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aSignedIn(ctx, "student"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx = contextWithToken(s, ctx)

	chapters := stepState.Request.([]*cpb.Chapter)
	ids := make([]string, 0, len(chapters))
	for _, chapter := range chapters {
		ids = append(ids, chapter.Info.Id)
	}
	filter := &cpb.CommonFilter{
		Ids: ids,
	}
	paging := &cpb.Paging{
		Limit: 5,
	}

	if err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		for {
			resp, err := epb.NewChapterReaderServiceClient(s.Conn).ListChapters(ctx, &epb.ListChaptersRequest{
				Filter: filter,
				Paging: paging,
			})
			if err != nil {
				return true, err
			}
			if len(resp.Items) == 0 {
				return false, nil
			}
			if len(resp.Items) > int(paging.Limit) {
				return attempt < 10, fmt.Errorf("unexpected total chapters: got: %d, want: %d", len(resp.Items), paging.Limit)
			}

			stepState.PaginatedChapters = append(stepState.PaginatedChapters, resp.Items)
			paging = resp.NextPage
		}
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnAListOfChapters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedChapters := stepState.Request.([]*cpb.Chapter)

	isValidChapter := func(chapter *cpb.Chapter, chapters []*cpb.Chapter) bool {
		for _, b := range chapters {
			if chapter.Info.Id == b.Info.Id {
				return true
			}
		}
		return false
	}

	var total int
	for _, chapters := range stepState.PaginatedChapters {
		if !sort.SliceIsSorted(chapters, func(i, j int) bool {
			return chapters[i].Info.DisplayOrder < chapters[j].Info.DisplayOrder
		}) {
			return StepStateToContext(ctx, stepState), errors.New("chapters are not sorted by display_order ASC")
		}

		for _, chapter := range chapters {
			if !isValidChapter(chapter, expectedChapters) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected chapter id: %q", chapter.Info.Id)
			}
		}

		total += len(chapters)
	}
	if total != len(expectedChapters) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total chapters: got %d, want: %d", total, len(expectedChapters))
	}
	return StepStateToContext(ctx, stepState), nil
}

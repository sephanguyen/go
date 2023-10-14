package common

import (
	"context"
	"fmt"
	"time"

	bob_constants "github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repositories "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgconn"
	"go.uber.org/multierr"
)

func (s *suite) aListOfValidChaptersInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < 10; i++ {
		c := new(bob_entities.Chapter)
		database.AllNullEntity(c)
		now := time.Now()
		id := fmt.Sprintf("chapter_id_%d", i)
		name := fmt.Sprintf("chapter_name_%d", i)
		c.ID.Set(id)
		c.Name.Set(name)
		if err := c.CreatedAt.Set(now); err != nil {
			return nil, err
		}
		if err := c.UpdatedAt.Set(now); err != nil {
			return nil, err
		}
		c.Country.Set(pb.COUNTRY_VN.String())
		c.Grade.Set(1)
		c.Subject.Set(pb.SUBJECT_CHEMISTRY.String())
		c.DisplayOrder.Set(1)
		c.SchoolID.Set(bob_constants.ManabieSchool)
		c.DeletedAt.Set(nil)

		_, err := database.Insert(ctx, c, s.BobDB.Exec)
		if e, ok := err.(*pgconn.PgError); ok && e.Code != "23505" {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.CurrentChapterIDs = append(stepState.CurrentChapterIDs, id)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAChapter(ctx context.Context) (context.Context, error) {
	chapter := &bob_entities.Chapter{}
	now := time.Now()
	database.AllNullEntity(chapter)
	stepState := StepStateFromContext(ctx)
	stepState.ChapterID = s.newID()
	multierr.Combine(
		chapter.ID.Set(stepState.ChapterID),
		chapter.Country.Set(pb.COUNTRY_VN),
		chapter.Name.Set(fmt.Sprintf("name-%s", stepState.ChapterID)),
		chapter.Grade.Set(12),
		chapter.SchoolID.Set(bob_constants.ManabieSchool),
		chapter.CurrentTopicDisplayOrder.Set(0),
		chapter.CreatedAt.Set(now),
		chapter.UpdatedAt.Set(now),
	)
	_, err := database.Insert(ctx, chapter, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateChapter(ctx context.Context, country, subject, bookID string, grade, schoolID int, chapterIDs []string) (*bob_entities.Chapter, error) {
	chapter1 := &bob_entities.Chapter{}
	database.AllNullEntity(chapter1)
	err := multierr.Combine(chapter1.ID.Set("book-chapter-"+s.newID()), chapter1.Name.Set("book-chapter-name-"+bookID), chapter1.Country.Set(country), chapter1.Subject.Set(subject), chapter1.Grade.Set(grade), chapter1.DisplayOrder.Set(1), chapter1.SchoolID.Set(schoolID), chapter1.UpdatedAt.Set(time.Now()), chapter1.CreatedAt.Set(time.Now()), chapter1.DeletedAt.Set(nil), chapter1.CurrentTopicDisplayOrder.Set(0))
	if err != nil {
		return nil, err
	}
	cmdTag, err := database.Insert(ctx, chapter1, s.BobDB.Exec)
	if err != nil {
		return nil, fmt.Errorf("database.Insert chapter: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		if err != nil {
			return nil, fmt.Errorf("database.Insert chapter: %w", bob_repositories.ErrUnAffected)
		}
	}
	chapterIDs = append(chapterIDs, chapter1.ID.String)
	for _, chapterID := range chapterIDs {
		bookChapter := &bob_entities.BookChapter{}
		database.AllNullEntity(bookChapter)
		err := multierr.Combine(bookChapter.BookID.Set(bookID), bookChapter.ChapterID.Set(chapterID), bookChapter.UpdatedAt.Set(time.Now()), bookChapter.CreatedAt.Set(time.Now()), bookChapter.DeletedAt.Set(nil))
		if err != nil {
			return nil, err
		}
		cmdTag, err := database.Insert(ctx, bookChapter, s.BobDB.Exec)
		if err != nil {
			return nil, fmt.Errorf("database.Insert book chapter: %w", err)
		}
		if cmdTag.RowsAffected() != 1 {
			if err != nil {
				return nil, fmt.Errorf("database.Insert book chapter: %w", repository.ErrUnAffected)
			}
		}
	}
	return chapter1, nil
}

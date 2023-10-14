package common

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	bob_repositories "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

func (s *suite) generateBook(ctx context.Context, courseID, country, subject string, grade, schoolID int, chapterIDs []string) error {
	now := time.Now()
	book := &entities.Book{}
	database.AllNullEntity(book)
	bookName := "book-name-course-id_" + courseID
	err := multierr.Combine(book.Country.Set(country), book.SchoolID.Set(schoolID), book.Subject.Set(subject), book.Grade.Set(grade), book.Name.Set(bookName), book.ID.Set(s.newID()), book.CreatedAt.Set(now), book.UpdatedAt.Set(now), book.CurrentChapterDisplayOrder.Set(0))
	if err != nil {
		return err
	}
	cmdTag, err := database.Insert(ctx, book, s.BobDB.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert book: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		if err != nil {
			return fmt.Errorf("database.Insert book: %w", bob_repositories.ErrUnAffected)
		}
	}
	courseBook := &entities.CoursesBooks{}
	database.AllNullEntity(courseBook)
	err = multierr.Combine(courseBook.CourseID.Set(courseID), courseBook.BookID.Set(book.ID.String), courseBook.CreatedAt.Set(now), courseBook.UpdatedAt.Set(now))
	if err != nil {
		return err
	}
	cmdTag, err = database.Insert(ctx, courseBook, s.BobDB.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert course book: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		if err != nil {
			return fmt.Errorf("database.Insert course book: %w", bob_repositories.ErrUnAffected)
		}
	}
	_, err = s.generateChapter(ctx, country, subject, book.ID.String, grade, schoolID, chapterIDs)
	if err != nil {
		return err
	}
	return nil
}

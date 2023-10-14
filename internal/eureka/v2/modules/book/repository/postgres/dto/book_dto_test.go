package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestToBookEntity(t *testing.T) {
	t.Parallel()
	now := time.Now()
	t.Run("convert to entity with deleted at", func(t *testing.T) {
		// arrange
		book := domain.Book{
			ID:                         "book-id-1",
			Name:                       "book-name-1",
			CreatedAt:                  now,
			UpdatedAt:                  now,
			DeletedAt:                  &now,
			CopiedFrom:                 "CopiedFrom",
			CurrentChapterDisplayOrder: 1,
			BookType:                   "type",
			IsV2:                       false,
		}
		dto := BookDto{
			ID:                         database.Text("book-id-1"),
			Name:                       database.Text("book-name-1"),
			CreatedAt:                  database.Timestamptz(now),
			UpdatedAt:                  database.Timestamptz(now),
			DeletedAt:                  database.Timestamptz(now),
			CopiedFrom:                 database.Text("CopiedFrom"),
			CurrentChapterDisplayOrder: database.Int4(1),
			BookType:                   database.Text("type"),
			IsV2:                       database.Bool(false),
		}

		// act
		res := dto.ToBookEntity()

		// assert
		assert.Equal(t, book, res)
	})

	t.Run("convert to entity with deleted at is nil", func(t *testing.T) {
		// arrange
		book := domain.Book{
			ID:                         "book-id-2",
			Name:                       "book-name-2",
			CreatedAt:                  now,
			UpdatedAt:                  now,
			DeletedAt:                  nil,
			CopiedFrom:                 "CopiedFrom",
			CurrentChapterDisplayOrder: 1,
			BookType:                   "type",
			IsV2:                       false,
		}
		dto := BookDto{
			ID:                         database.Text("book-id-2"),
			Name:                       database.Text("book-name-2"),
			CreatedAt:                  database.Timestamptz(now),
			UpdatedAt:                  database.Timestamptz(now),
			DeletedAt:                  pgtype.Timestamptz{Status: pgtype.Null},
			CopiedFrom:                 database.Text("CopiedFrom"),
			CurrentChapterDisplayOrder: database.Int4(1),
			BookType:                   database.Text("type"),
			IsV2:                       database.Bool(false),
		}

		// act
		res := dto.ToBookEntity()

		// assert
		assert.Equal(t, book, res)
	})
}

func TestNewBookDtoFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Now()

	t.Run("convert to dto successfully", func(t *testing.T) {
		// arrange
		book := domain.Book{
			ID:                         "book-id-3",
			Name:                       "book-name-3",
			CreatedAt:                  now,
			UpdatedAt:                  now,
			DeletedAt:                  &now,
			CopiedFrom:                 "CopiedFrom3",
			CurrentChapterDisplayOrder: 1,
			BookType:                   "type3",
			IsV2:                       false,
		}
		dto := BookDto{
			ID:                         database.Text("book-id-3"),
			Name:                       database.Text("book-name-3"),
			CreatedAt:                  database.Timestamptz(now),
			UpdatedAt:                  database.Timestamptz(now),
			DeletedAt:                  database.Timestamptz(now),
			CopiedFrom:                 database.Text("CopiedFrom3"),
			CurrentChapterDisplayOrder: database.Int4(1),
			BookType:                   database.Text("type3"),
			IsV2:                       database.Bool(false),
		}

		// act
		res, err := NewBookDtoFromEntity(book)

		// assert
		assert.Equal(t, dto, res)
		assert.Nil(t, err)
	})

	t.Run("convert to dto with nil deleted at", func(t *testing.T) {
		// arrange
		book := domain.Book{
			ID:                         "book-id-4",
			Name:                       "book-name-4",
			CreatedAt:                  now,
			UpdatedAt:                  now,
			DeletedAt:                  nil,
			CopiedFrom:                 "CopiedFrom4",
			CurrentChapterDisplayOrder: 1,
			BookType:                   "type4",
			IsV2:                       false,
		}
		dto := BookDto{
			ID:                         database.Text("book-id-4"),
			Name:                       database.Text("book-name-4"),
			CreatedAt:                  database.Timestamptz(now),
			UpdatedAt:                  database.Timestamptz(now),
			DeletedAt:                  pgtype.Timestamptz{Status: pgtype.Null},
			CopiedFrom:                 database.Text("CopiedFrom4"),
			CurrentChapterDisplayOrder: database.Int4(1),
			BookType:                   database.Text("type4"),
			IsV2:                       database.Bool(false),
		}

		// act
		res, err := NewBookDtoFromEntity(book)

		// assert
		assert.Equal(t, dto, res)
		assert.Nil(t, err)
	})
}

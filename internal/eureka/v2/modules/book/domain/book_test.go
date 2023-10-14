package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBook_RemoveUnpublishedContent(t *testing.T) {
	t.Run("Remove all unpublished content recursively", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, _, unPub := GenFakeChapters(10, 0)
		book := Book{
			ID:       "Book ID",
			Name:     "Book name",
			Chapters: unPub,
		}
		expectedBook := Book{
			ID:       book.ID,
			Name:     book.Name,
			Chapters: []Chapter{},
		}
		// act
		book.RemoveUnpublishedContent()

		// assert
		assert.Equal(t, expectedBook, book)
	})

	t.Run("Remove a set of unpublished content recursively", func(t *testing.T) {
		t.Parallel()
		// arrange
		all, published, _ := GenFakeChapters(10, 5)
		book := Book{
			ID:       "Book ID",
			Name:     "Book name",
			Chapters: all,
		}
		expectedBook := Book{
			ID:       book.ID,
			Name:     book.Name,
			Chapters: published,
		}
		// act
		book.RemoveUnpublishedContent()

		// assert
		assert.Equal(t, expectedBook, book)
	})

	t.Run("Remove nothing cause all of them are published recursively", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, published, _ := GenFakeChapters(10, 5)
		book := Book{
			ID:       "Book ID",
			Name:     "Book name",
			Chapters: published,
		}
		expectedBook := Book{
			ID:       book.ID,
			Name:     book.Name,
			Chapters: published,
		}
		// act
		book.RemoveUnpublishedContent()

		// assert
		assert.Equal(t, expectedBook, book)
	})
}

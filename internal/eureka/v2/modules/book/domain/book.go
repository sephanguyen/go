package domain

import (
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2/common"
)

type Book struct {
	ID                         string
	Name                       string
	CopiedFrom                 string
	CurrentChapterDisplayOrder int
	BookType                   string
	IsV2                       bool

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt *time.Time

	// virtual properties
	Chapters []Chapter
}

type BookHierarchyFlatten struct {
	BookID             string
	ChapterID          string
	TopicID            string
	LearningMaterialID string
}

func NewBook(bookID, name string) Book {
	return Book{
		ID:                         bookID,
		Name:                       name,
		CurrentChapterDisplayOrder: 0,
		BookType:                   cpb.BookType_BOOK_TYPE_GENERAL.String(),
		IsV2:                       true,
	}
}

// RemoveUnpublishedContent removes all unpublished content recursively
//
// Published book is book that contains published chapters.
//
// Published chapter is chapter that contains published topics.
//
// Published topic is topic that contains published learning materials.
func (b *Book) RemoveUnpublishedContent() {
	publishedChapters := make([]Chapter, 0, len(b.Chapters))

	for i := 0; i < len(b.Chapters); i++ {
		chapter := b.Chapters[i]
		chapter.RemoveUnpublishedTopics()
		if len(chapter.Topics) > 0 {
			publishedChapters = append(publishedChapters, chapter)
		}
	}

	b.Chapters = publishedChapters
}

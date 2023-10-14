package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type BookDto struct {
	ID                         pgtype.Text
	Name                       pgtype.Text
	UpdatedAt                  pgtype.Timestamptz
	CreatedAt                  pgtype.Timestamptz
	DeletedAt                  pgtype.Timestamptz
	CopiedFrom                 pgtype.Text
	CurrentChapterDisplayOrder pgtype.Int4
	BookType                   pgtype.Text
	IsV2                       pgtype.Bool
}

func (b BookDto) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "name", "updated_at", "created_at", "deleted_at", "copied_from", "current_chapter_display_order", "book_type", "is_v2"}
	values = []interface{}{&b.ID, &b.Name, &b.UpdatedAt, &b.CreatedAt, &b.DeletedAt, &b.CopiedFrom, &b.CurrentChapterDisplayOrder, &b.BookType, &b.IsV2}
	return
}

func (b BookDto) TableName() string {
	return "books"
}

func (b BookDto) ToBookEntity() domain.Book {
	book := domain.Book{
		ID:                         b.ID.String,
		Name:                       b.Name.String,
		CreatedAt:                  b.CreatedAt.Time,
		UpdatedAt:                  b.UpdatedAt.Time,
		CopiedFrom:                 b.CopiedFrom.String,
		CurrentChapterDisplayOrder: int(b.CurrentChapterDisplayOrder.Int),
		BookType:                   b.BookType.String,
		IsV2:                       b.IsV2.Bool,
	}
	if b.DeletedAt.Status == pgtype.Present {
		book.DeletedAt = &b.DeletedAt.Time
	}
	return book
}

func NewBookDtoFromEntity(book domain.Book) (BookDto, error) {
	dto := BookDto{}
	database.AllNullEntity(&dto)
	err := multierr.Combine(
		dto.ID.Set(book.ID),
		dto.Name.Set(book.Name),
		dto.CopiedFrom.Set(book.CopiedFrom),
		dto.CurrentChapterDisplayOrder.Set(book.CurrentChapterDisplayOrder),
		dto.BookType.Set(book.BookType),
		dto.IsV2.Set(book.IsV2),

		dto.CreatedAt.Set(book.CreatedAt),
		dto.UpdatedAt.Set(book.UpdatedAt),
		dto.DeletedAt.Set(book.DeletedAt),
	)

	return dto, err
}

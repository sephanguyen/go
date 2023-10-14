package database

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

type BookChapter struct {
	BookID    pgtype.Text `sql:"book_id"`
	ChapterID pgtype.Text `sql:"chapter_id"`
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *BookChapter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "chapter_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.BookID, &c.ChapterID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*BookChapter) TableName() string {
	return "books_chapters"
}

const bookChaptersJsonString = `{
		"schema": [
			{
				"column_name": "book_id",
				"data_type": "text",
				"column_default": null
			},
			{
				"column_name": "chapter_id",
				"data_type": "text",
				"column_default": null
			},
			{
				"column_name": "updated_at",
				"data_type": "timestamp with time zone",
				"column_default": null
			},
			{
				"column_name": "created_at",
				"data_type": "timestamp with time zone",
				"column_default": null
			},
			{
				"column_name": "deleted_at",
				"data_type": "timestamp with time zone",
				"column_default": null
			},
			{
				"column_name": "resource_path",
				"data_type": "text",
				"column_default": "autofillresourcepath()"
			}
		],
		"policies": [
			{
				"tablename": "books_chapters",
				"policyname": "rls_books_chapters",
				"qual": "permission_check(resource_path, (books_chapters.*)::text)",
				"with_check": "permission_check(resource_path, (books_chapters.*)::text)",
				"relrowsecurity": false,
				"relforcerowsecurity": false
			},
			{
				"tablename": "books_chapters",
				"policyname": "rls_books_chapters",
				"qual": "permission_check(resource_path, (books_chapters.*)::text)",
				"with_check": "permission_check(resource_path, (books_chapters.*)::text)",
				"relrowsecurity": false,
				"relforcerowsecurity": false
			}
		],
		"table_name": "books_chapters"
	}`

func TestVerifyEntity(t *testing.T) {
	t.Parallel()
	t.Run("test with require enable RLS Service and valid resource_path default value", func(t *testing.T) {
		t.Parallel()
		var service = &SchemaVerifier{
			Service: "bob",
		}
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NoError(t, err)
	})
	t.Run("test with require enable RLS Service and invalid resource_path default value", func(t *testing.T) {
		t.Parallel()
		var service = &SchemaVerifier{
			Service: "bob",
		}
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Schema[5].ColumnDefault = pgtype.Text{
			String: "something different with autofillresourcepath()",
		}

		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NotNil(t, err)
	})
	t.Run("test with ignore RLS Service", func(t *testing.T) {
		t.Parallel()
		var service = &SchemaVerifier{
			Service: "zeus",
		}
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NoError(t, err)

		bookChaptersSchema.Schema[5].ColumnDefault = pgtype.Text{
			String: "something different with autofillresourcepath()",
		}

		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NoError(t, err)
	})

	t.Run("test with table without resource_path field", func(t *testing.T) {
		var service = &SchemaVerifier{
			Service: "bob",
		}

		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NoError(t, err)

		bookChaptersSchema.Schema = append(bookChaptersSchema.Schema[:5], bookChaptersSchema.Schema[6:]...)
		err = service.VerifyEntity(bookChaptersSchema, &BookChapter{})
		assert.NotNil(t, err)
	})
}

package database

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestVerifyAllTableWithRLS(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.Nil(VerifyAllTableWithRLS())
}

func TestRLSTableCheck(t *testing.T) {
	t.Run("test with require enable RLS table and valid RLS Policy data", func(t *testing.T) {
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].Qual = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[0].WithCheck = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[0].Permissive = pgtype.Text{String: "PERMISSIVE"}
		bookChaptersSchema.Policies[0].Roles.Set([]string{"public", "bob"})

		bookChaptersSchema.Policies[1].PolicyName = pgtype.Text{String: "rls_books_chapters_restrictive"}
		bookChaptersSchema.Policies[1].Relforcerowsecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[1].RelrowSecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[1].Qual = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[1].WithCheck = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[1].Permissive = pgtype.Text{String: "RESTRICTIVE"}
		bookChaptersSchema.Policies[1].Roles.Set([]string{"public", "bob"})

		err = tableCheck("bob", bookChaptersSchema)
		assert.NoError(t, err)
	})

	t.Run("test with require enable RLS table and invalid policy name format", func(t *testing.T) {
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Policies[0].PolicyName = pgtype.Text{String: "invalid policy name format"}
		bookChaptersSchema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].Qual = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[0].WithCheck = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}
		bookChaptersSchema.Policies[0].Permissive = pgtype.Text{String: "PERMISSIVE"}
		bookChaptersSchema.Policies[0].Roles.Set([]string{"public", "bob"})

		err = tableCheck("bob", bookChaptersSchema)
		assert.Error(t, err)
		assert.Equal(t, "policy name is not in format rls_books_chapters or rls_books_chapters_restrictive for table books_chapters in service bob", err.Error())
	})

	t.Run("test with require enable RLS table and relrowsecurity is false", func(t *testing.T) {
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: false}

		err = tableCheck("bob", bookChaptersSchema)
		assert.Error(t, err)
		assert.Equal(t, "row security is not enable for table books_chapters in service bob", err.Error())
	})

	t.Run("test with require enable RLS table and relforcerowsecurity is false", func(t *testing.T) {
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: false}
		bookChaptersSchema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: true}

		err = tableCheck("bob", bookChaptersSchema)
		assert.Error(t, err)
		assert.Equal(t, "please force row level security for table books_chapters in service bob", err.Error())
	})

	t.Run("test with require permission_check function in policy", func(t *testing.T) {
		var bookChaptersSchema = &tableSchema{}
		err := json.Unmarshal([]byte(bookChaptersJsonString), &bookChaptersSchema)
		assert.NoError(t, err)

		bookChaptersSchema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: true}
		bookChaptersSchema.Policies[0].WithCheck = pgtype.Text{String: "permission_check(resource_path, 'books_chapters'::text)"}

		err = tableCheck("bob", bookChaptersSchema)
		assert.Error(t, err)
		assert.Equal(t, "function permission_check is not in policy for table books_chapters in service bob. Please change to permission_check(resource_path, 'books_chapters'::text)", err.Error())
	})
}

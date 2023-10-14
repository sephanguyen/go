package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePlaceholders(t *testing.T) {
	t.Parallel()
	t.Run("success returns empty string", func(t *testing.T) {
		t.Parallel()
		cases := []int{0, -1, -123456}
		for _, v := range cases {
			actual := GeneratePlaceholders(v)
			assert.Empty(t, actual)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		actual := GeneratePlaceholders(4)
		assert.Exactly(t, "$1, $2, $3, $4", actual)
	})
}

func TestGeneratePlaceholdersWithFirstIndex(t *testing.T) {
	t.Parallel()

	t.Run("success returns empty string", func(t *testing.T) {
		t.Parallel()
		cases := []int{0, -1, -123456}
		for _, n := range cases {
			actual := GeneratePlaceholdersWithFirstIndex(2, n)
			assert.Empty(t, actual)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		actual := GeneratePlaceholdersWithFirstIndex(1, 4)
		assert.Exactly(t, "$1, $2, $3, $4", actual)

		actual = GeneratePlaceholdersWithFirstIndex(2, 1)
		assert.Exactly(t, "$2", actual)

		actual = GeneratePlaceholdersWithFirstIndex(0, 3)
		assert.Exactly(t, "$1, $2, $3", actual)
	})
}

func TestTrimFieldEntityTableName(t *testing.T) {
	t.Parallel()
	ent, err := newExampleEntity()
	assert.NoError(t, err)
	tfe := TrimFieldEntity{E: ent, N: 0}
	assert.Exactly(t, "test_table", tfe.TableName())
}

func TestTrimFieldEntityFieldMap(t *testing.T) {
	t.Parallel()
	ent, err := newExampleEntity()
	assert.NoError(t, err)

	t.Run("success with N <= 0", func(t *testing.T) {
		t.Parallel()
		expectedFields := []string{
			"bool",
			"int2",
			"int4",
			"int4array",
			"text",
			"textarray",
			"varchar",
			"timestamptz",
			"jsonb",
		}
		expectedValues := []interface{}{
			&ent.Bool,
			&ent.Int2,
			&ent.Int4,
			&ent.Int4Array,
			&ent.Text,
			&ent.TextArray,
			&ent.Varchar,
			&ent.Timestamptz,
			&ent.JSONB,
		}

		for n := -1; n <= 0; n++ {
			tfe := TrimFieldEntity{E: ent, N: n}
			actualFields, actualValues := tfe.FieldMap()
			assert.Exactly(t, expectedFields, actualFields)
			assert.Exactly(t, expectedValues, actualValues)
		}
	})

	t.Run("success with N = 5", func(t *testing.T) {
		t.Parallel()
		tfe := TrimFieldEntity{E: ent, N: 5}
		actualFields, actualValues := tfe.FieldMap()
		expectedFields := []string{
			"textarray",
			"varchar",
			"timestamptz",
			"jsonb",
		}
		expectedValues := []interface{}{
			&ent.TextArray,
			&ent.Varchar,
			&ent.Timestamptz,
			&ent.JSONB,
		}

		assert.Exactly(t, expectedFields, actualFields)
		assert.Exactly(t, expectedValues, actualValues)
	})

	t.Run("empty return with N > len(fields)", func(t *testing.T) {
		t.Parallel()
		tfe := TrimFieldEntity{E: ent, N: 99999}
		actualFields, actualValues := tfe.FieldMap()
		assert.Exactly(t, []string{}, actualFields)
		assert.Exactly(t, []interface{}{}, actualValues)
	})
}

func TestGenerateUpdatePlaceholders(t *testing.T) {
	t.Parallel()
	t.Run("success empty", func(t *testing.T) {
		t.Parallel()
		actual := generateUpdatePlaceholders([]string{})
		assert.Empty(t, actual)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		actual := generateUpdatePlaceholders([]string{"abc", "defg", "xyz1234"})
		assert.Exactly(t, "abc = $1, defg = $2, xyz1234 = $3", actual)
	})
}

func Test_GenerateUpdatePlaceholders(t *testing.T) {
	t.Parallel()
	t.Run("success empty", func(t *testing.T) {
		t.Parallel()
		actual := GenerateUpdatePlaceholders([]string{}, 1)
		assert.Empty(t, actual)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		actual := GenerateUpdatePlaceholders([]string{"abc", "defg", "xyz1234"}, 1)
		assert.Exactly(t, "abc = $1, defg = $2, xyz1234 = $3", actual)

		actual = GenerateUpdatePlaceholders([]string{"abc"}, 3)
		assert.Exactly(t, "abc = $3", actual)
	})
}

func TestFindColumn(t *testing.T) {
	t.Parallel()
	t.Run("found", func(t *testing.T) {
		t.Parallel()
		actual := FindColumn([]string{"col1", "col2", "col3"}, "col2")
		assert.Exactly(t, 1, actual)
	})

	t.Run("cannot find", func(t *testing.T) {
		t.Parallel()
		actual := FindColumn([]string{"col1", "col2", "col3"}, "col999")
		assert.Exactly(t, -1, actual)
	})
}

func TestAddPagingQuery(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		query, args := AddPagingQuery("SOME $1 SQL QUERY $2 HERE", 123, 4, "value1", "value2")
		assert.Exactly(t, "SOME $1 SQL QUERY $2 HERE LIMIT $3 OFFSET $4", query)
		assert.Exactly(t, []interface{}{"value1", "value2", int32(123), int32(123 * 3)}, args)
	})

	t.Run("success limit=0", func(t *testing.T) {
		t.Parallel()
		query, args := AddPagingQuery("SOME $1 SQL QUERY $2 HERE", 0, 1, "value1", "value2")
		assert.Exactly(t, "SOME $1 SQL QUERY $2 HERE LIMIT $3 OFFSET $4", query)
		assert.Exactly(t, []interface{}{"value1", "value2", int32(10), int32(0)}, args)
	})

	t.Run("ignored due to page=0", func(t *testing.T) {
		t.Parallel()
		query, args := AddPagingQuery("SOME $1 SQL QUERY $2 HERE", 123, 0, "value1", "value2")
		assert.Exactly(t, "SOME $1 SQL QUERY $2 HERE", query)
		assert.Exactly(t, []interface{}{"value1", "value2"}, args)
	})
}

func TestCompositeKeysPlaceHolders(t *testing.T) {
	t.Parallel()
	t.Run("example 1", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]string{
			{"a", "b"},
			{"c", "d"},
		}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			return []interface{}{keys[i][0], keys[i][1]}
		})

		assert.Equal(tt, "($1, $2), ($3, $4)", str)
		assert.Equal(tt, []interface{}{"a", "b", "c", "d"}, args)
	})

	t.Run("example 2", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]string{}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			return []interface{}{keys[i][0], keys[i][1]}
		})

		assert.Equal(tt, "", str)
		assert.Nil(tt, args)
	})

	t.Run("example 3", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]string{
			{"a", "b"},
			{},
			{"c", "d"},
		}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			args := []interface{}{}
			for _, v := range keys[i] {
				args = append(args, v)
			}
			return args
		})

		assert.Equal(tt, "($1, $2), (), ($3, $4)", str)
		assert.Equal(tt, []interface{}{"a", "b", "c", "d"}, args)
	})

	t.Run("example 4", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]interface{}{
			{"a", "b"},
			{1},
			{"c", "d"},
		}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			return keys[i]
		})

		assert.Equal(tt, "($1, $2), ($3), ($4, $5)", str)
		assert.Equal(tt, []interface{}{"a", "b", 1, "c", "d"}, args)
	})

	t.Run("example 5", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]interface{}{
			{"a"},
			{1},
			{},
			{"c"},
		}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			return keys[i]
		})

		assert.Equal(tt, "($1), ($2), (), ($3)", str)
		assert.Equal(tt, []interface{}{"a", 1, "c"}, args)
	})

	t.Run("example 6", func(tt *testing.T) {
		tt.Parallel()
		keys := [][]interface{}{
			{"a"},
			{1},
			{},
			{"c"},
			{},
			{"e", "f", 2},
		}
		str, args := CompositeKeysPlaceHolders(len(keys), func(i int) []interface{} {
			return keys[i]
		})

		assert.Equal(tt, "($1), ($2), (), ($3), (), ($4, $5, $6)", str)
		assert.Equal(tt, []interface{}{"a", 1, "c", "e", "f", 2}, args)
	})
}

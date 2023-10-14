package database

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

var sharedTime = time.Now()

// exampleUserDefinedEntity is an example of a user-defined entity. Some real instances
// are Bob's entities.City or entities.District.
type exampleUserDefinedEntity struct {
	Bool pgtype.Bool
	Int2 pgtype.Int2
}

// exampleEntity contains fields that are frequently used throughout manabie services.
// There are complex fields such as City or District that have not been added.
type exampleEntity struct {
	// Basic pgtype types which can be set
	Bool        pgtype.Bool
	Int2        pgtype.Int2
	Int4        pgtype.Int4
	Int4Array   pgtype.Int4Array
	Text        pgtype.Text
	TextArray   pgtype.TextArray
	Varchar     pgtype.Varchar
	Timestamptz pgtype.Timestamptz
	JSONB       pgtype.JSONB

	// Go type which does not implement setter interface
	NonPGType string

	// User-defined types which do not implement setter interface
	UserDefined exampleUserDefinedEntity
}

func newExampleEntity() (*exampleEntity, error) {
	e := &exampleEntity{
		Bool:      Bool(true),
		Int2:      Int2(123),
		Int4:      Int4(456),
		Int4Array: Int4Array([]int32{1, 2, 3}),
		Text:      Text("a string"),
		TextArray: TextArray([]string{"a", "slice", "of", "string"}),
		Varchar:   Varchar("another string"),
		JSONB:     JSONB([]byte(`{}`)),

		NonPGType:   "cannot be set",
		UserDefined: exampleUserDefinedEntity{Bool: Bool(true), Int2: Int2(123)},
	}

	ts := Timestamptz(sharedTime)
	e.Timestamptz = ts
	return e, nil
}

func (e *exampleEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}
	values = []interface{}{&e.Bool, &e.Int2, &e.Int4, &e.Int4Array, &e.Text, &e.TextArray, &e.Varchar, &e.Timestamptz, &e.JSONB}
	return
}

func (*exampleEntity) TableName() string {
	return "test_table"
}

type missingFieldNameEntity exampleEntity

func (me *missingFieldNameEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz"} // fail due to missing "jsonb"
	values = []interface{}{&me.Bool, &me.Int2, &me.Int4, &me.Int4Array, &me.Text, &me.TextArray, &me.Varchar, &me.Timestamptz, &me.JSONB}
	return
}

func (*missingFieldNameEntity) TableName() string {
	return "test_table"
}

type missingFieldValueEntity exampleEntity

func (me *missingFieldValueEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}
	values = []interface{}{&me.Bool, &me.Int2, &me.Int4, &me.Int4Array, &me.Text, &me.TextArray, &me.Varchar} // fail due to missing &me.Timestamptz and &me.JSONB
	return
}

func (*missingFieldValueEntity) TableName() string {
	return "test_table"
}

type nonPtrValueEntity exampleEntity

func (ne *nonPtrValueEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}
	values = []interface{}{&ne.Bool, &ne.Int2, &ne.Int4, &ne.Int4Array, &ne.Text, &ne.TextArray, &ne.Varchar, &ne.Timestamptz, ne.JSONB}
	return
}

func (*nonPtrValueEntity) TableName() string {
	return "test_table"
}

type missingTableNameEntity exampleEntity

func (me *missingTableNameEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}
	values = []interface{}{&me.Bool, &me.Int2, &me.Int4, &me.Int4Array, &me.Text, &me.TextArray, &me.Varchar, &me.Timestamptz, &me.JSONB}
	return
}

func (*missingTableNameEntity) TableName() string {
	return "" // return empty table name
}

type duplicatedFieldNameEntity exampleEntity

func (de *duplicatedFieldNameEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "bool", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"} // duplicated "bool"
	values = []interface{}{&de.Bool, &de.Int2, &de.Int4, &de.Int4Array, &de.Text, &de.TextArray, &de.Varchar, &de.Timestamptz, &de.JSONB}
	return
}

func (*duplicatedFieldNameEntity) TableName() string {
	return "test_table"
}

type duplicatedFieldValueEntity exampleEntity

func (de *duplicatedFieldValueEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}
	values = []interface{}{&de.Bool, &de.Int4, &de.Int4, &de.Int4Array, &de.Text, &de.TextArray, &de.Varchar, &de.Timestamptz, &de.JSONB} // duplicated &de.Int4
	return
}

func (*duplicatedFieldValueEntity) TableName() string {
	return "test_table"
}

// exampleEntities, nonSliceEntities, and errornousAddEntities implement different cases to test database.Entities.
type exampleEntities []*exampleEntity

func (u *exampleEntities) Add() Entity {
	e := &exampleEntity{}
	*u = append(*u, e)
	return e
}

type nonSliceEntities exampleEntity

func (*nonSliceEntities) Add() Entity {
	e := &exampleEntity{}
	return e
}

type errornousAddEntities []*exampleEntity

func (*errornousAddEntities) Add() Entity {
	e := &exampleEntity{}
	// missing the append line
	return e
}

func TestCheckEntityDefinition(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assert.NoError(CheckEntityDefinition(&exampleEntity{}))
	})

	// errornous cases
	testcases := []struct {
		name      string
		ent       Entity
		errString string
	}{
		{"missing field name", &missingFieldNameEntity{}, `missingFieldNameEntity.FieldMap() returned 8 field names, expected 9`},
		{"missing field value", &missingFieldValueEntity{}, `missingFieldValueEntity.FieldMap() returned 7 field values, expected 9`},
		{"non-pointer field value", &nonPtrValueEntity{}, "field jsonb of nonPtrValueEntity is not a pointer"},
		{"missing table name", &missingTableNameEntity{}, `missingTableNameEntity.TableName() returned empty`},
		{"duplicated field name", &duplicatedFieldNameEntity{}, `duplicated field name "bool" (index 0 and 1) in entity duplicatedFieldNameEntity`},
		{"duplicated field value", &duplicatedFieldValueEntity{}, `duplicated pointer value (index 1 and 2) in entity duplicatedFieldValueEntity`},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.EqualError(
				CheckEntityDefinition(tc.ent),
				tc.errString,
			)
		})
	}
}

func TestCheckEntitiesDefinition(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, CheckEntitiesDefinition(&exampleEntities{}))
	})

	t.Run("is not a slice type", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(
			t,
			CheckEntitiesDefinition(&nonSliceEntities{}),
			"database.nonSliceEntities's underlying type must be a slice",
		)
	})

	t.Run("errornous Add implementation", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(
			t,
			CheckEntitiesDefinition(&errornousAddEntities{}),
			`database.errornousAddEntities.Add fails to add a new element (oldLen: 0, newLen: 0)`,
		)
	})
}

func TestAllNullEntity(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ent, err := newExampleEntity()
	assert.NoError(err)

	AllNullEntity(ent)
	assert.Equal(pgtype.Null, ent.Bool.Status)
	assert.Equal(pgtype.Null, ent.Int2.Status)
	assert.Equal(pgtype.Null, ent.Int4.Status)
	assert.Equal(pgtype.Null, ent.Int4Array.Status)
	assert.Equal(pgtype.Null, ent.Text.Status)
	assert.Equal(pgtype.Null, ent.TextArray.Status)
	assert.Equal(pgtype.Null, ent.Varchar.Status)
	assert.Equal(pgtype.Null, ent.Timestamptz.Status)
	assert.Equal(pgtype.Null, ent.JSONB.Status)

	assert.Equal("cannot be set", ent.NonPGType)
	assert.Equal(exampleUserDefinedEntity{Bool: Bool(true), Int2: Int2(123)}, ent.UserDefined)
}

func TestGetFieldNames(t *testing.T) {
	t.Parallel()
	expected := []string{"bool", "int2", "int4", "int4array", "text", "textarray", "varchar", "timestamptz", "jsonb"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, expected, GetFieldNames(&exampleEntity{}))
	})
}

func TestGetScanFields(t *testing.T) {
	t.Parallel()
	ent, err := newExampleEntity()
	assert.NoError(t, err)

	t.Run("success all valid fields", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		reqlist := GetFieldNames(ent)
		result := GetScanFields(ent, reqlist)

		// Create a new object using the same function to compare
		expected, err := newExampleEntity()
		assert.NoError(err)
		assert.Len(result, 9)
		assert.Exactly(&expected.Bool, result[0])
		assert.Exactly(&expected.Int2, result[1])
		assert.Exactly(&expected.Int4, result[2])
		assert.Exactly(&expected.Int4Array, result[3])
		assert.Exactly(&expected.Text, result[4])
		assert.Exactly(&expected.TextArray, result[5])
		assert.Exactly(&expected.Varchar, result[6])
		assert.Exactly(&expected.Timestamptz, result[7])
		assert.Exactly(&expected.JSONB, result[8])
	})

	t.Run("success invalid fields", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		reqlist := []string{"some", "fields", "that", "don't", "exist"}
		result := GetScanFields(ent, reqlist)
		assert.Empty(result)
	})
}

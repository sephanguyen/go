package database

import (
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

var testrec *SchemaVerifier

func TestMain(m *testing.M) {
	// Change migration directory to mock/testing/testdata/test/migrations
	migrationLoc = "../../../mock/testing/testdata/test/migrations"

	var err error
	testrec, err = NewSchemaVerifier("test")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

// TODO: Check if pgtype package has any built-in mechanism to check these pgtypes.
func TestTypeMap(t *testing.T) {
	assert := assert.New(t)
	assert.True(typemap[reflect.TypeOf(&pgtype.Bool{})]("boolean"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Date{})]("date"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Float4{})]("real"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Int2{})]("smallint"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Int2Array{})]("ARRAY"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Int4{})]("integer"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Int4Array{})]("ARRAY"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Int8{})]("bigint"))
	assert.True(typemap[reflect.TypeOf(&pgtype.JSON{})]("json"))
	assert.True(typemap[reflect.TypeOf(&pgtype.JSONB{})]("jsonb"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Numeric{})]("numeric"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Point{})]("point"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Text{})]("text"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Text{})]("character varying"))
	assert.True(typemap[reflect.TypeOf(&pgtype.TextArray{})]("ARRAY"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Timestamptz{})]("timestamp with time zone"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Varchar{})]("text"))
	assert.True(typemap[reflect.TypeOf(&pgtype.Varchar{})]("character varying"))
}

func TestEntityDefinitions(t *testing.T) {
	CheckEntityDefinition(&fieldSchema{})
	CheckEntityDefinition(&table{})
	CheckEntitiesDefinition(&fieldSchemas{})
	CheckEntitiesDefinition(&tables{})
	CheckEntityDefinition(&tablePolicy{})
}

// testentity is taken from bob_entities.Book.
type testentity struct {
	ID pgtype.Text
}

func (c *testentity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id"}
	values = []interface{}{&c.ID}
	return
}

func (*testentity) TableName() string {
	return "books"
}

type testentitymissingfield struct {
	ID           pgtype.Text
	MissingField pgtype.Text
}

func (c *testentitymissingfield) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "missing_field"}
	values = []interface{}{&c.ID, &c.MissingField}
	return
}

func (*testentitymissingfield) TableName() string {
	return "books"
}

type testentitywrongtablename struct {
	ID pgtype.Text
}

func (c *testentitywrongtablename) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id"}
	values = []interface{}{&c.ID}
	return
}

func (*testentitywrongtablename) TableName() string {
	return "books2"
}

type testentitywrongtype struct {
	ID pgtype.Int4
}

func (c *testentitywrongtype) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id"}
	values = []interface{}{&c.ID}
	return
}

func (*testentitywrongtype) TableName() string {
	return "books"
}

func TestSchemaVerifier_VerifyEntity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ent := &testentity{}
		CheckEntityDefinition(ent)
		assert.NoError(t, testrec.Verify(ent))
	})

	testcases := []struct {
		name string
		ent  Entity
	}{
		{name: "error missing field", ent: &testentitymissingfield{}},
		{name: "error mismatched table name", ent: &testentitywrongtablename{}},
		{name: "error mismatched type", ent: &testentitywrongtype{}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			CheckEntityDefinition(tc.ent)
			assert.Error(t, testrec.Verify(tc.ent))
		})
	}
}

func TestFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filePath string
		err      bool
	}{
		{
			name:     "version not number",
			filePath: "../version_migrate.up.sql",
			err:      true,
		},
		{
			name:     "version contain number",
			filePath: "../111d1_migrate.up.sql",
			err:      true,
		},
		{
			name:     "version range not valid",
			filePath: "../11111_migrate.up.sql",
			err:      true,
		},
		{
			name:     "file name not valid",
			filePath: "../1111_migration.up.sql",
			err:      true,
		},
		{
			name:     "happy case",
			filePath: "../1111_migrate.up.sql",
			err:      false,
		},
		{
			name:     "happy case",
			filePath: "../9999_migrate.up.sql",
			err:      false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := getVersion(tc.filePath); err != nil {
				assert.True(t, tc.err)
			} else {
				assert.False(t, tc.err)
			}
		})
	}
}

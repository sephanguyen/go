package exporter

import (
	"errors"
	"math/big"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/stretchr/testify/assert"
)

// sample
type TEntity struct {
	SomeName string
	Path     string
	Age      int
	Deleted  bool
}
type TEntity2 struct {
	SomeName string
	Path     string
	Age      float64
	Deleted  bool
}

func (t *TEntity) TableName() string {
	return "entity_table_name"
}

func (t *TEntity2) TableName() string {
	return "entity2_table_name"
}

func (e *TEntity) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"some_name",
		"path",
		"age",
		"deleted",
	}
	values = []interface{}{
		&e.SomeName,
		&e.Path,
		&e.Age,
		&e.Deleted,
	}
	return fields, values
}

func (e *TEntity2) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"some_name",
		"path",
		"age",
		"deleted",
	}
	values = []interface{}{
		e.SomeName,
		e.Path,
		e.Age,
		e.Deleted,
	}
	return fields, values
}

func TestExportBatch(t *testing.T) {
	t.Parallel()

	n1 := "Name 1 e1"
	n2 := "Name 2 e1"
	n3 := "Name 3 e1"
	p1 := "Path 1 e1"
	p2 := "Path 2 e1"
	p3 := "Path 3 e1"
	n1_2 := "Name 1 e2"
	n2_2 := "Name 2 e2"
	n3_2 := "Name 3 e2"
	p1_2 := "Path 1 e2"
	p2_2 := "Path 2 e2"
	p3_2 := "Path 3 e2"

	e1 := []database.Entity{
		&TEntity{
			SomeName: n1,
			Path:     p1,
			Age:      1,
			Deleted:  true,
		},
		&TEntity{
			SomeName: n2,
			Path:     p2,
			Age:      2,
			Deleted:  false,
		},
		&TEntity{
			SomeName: n3,
			Path:     p3,
			Age:      3,
			Deleted:  false,
		},
	}
	e2 := []database.Entity{
		&TEntity2{
			SomeName: n1_2,
			Path:     p1_2,
			Age:      10.55,
			Deleted:  true,
		},
		&TEntity2{
			SomeName: n2_2,
			Path:     p2_2,
			Age:      2.3333,
			Deleted:  true,
		},
		&TEntity2{
			SomeName: n3_2,
			Path:     p3_2,
			Age:      3.444,
			Deleted:  false,
		},
	}

	t.Run("should return error if column map is empty", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{}

		// act
		data, err := ExportBatch(e1, colMap)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, errors.New("column map should not be empty"))
	})

	t.Run("should return error if DBColumn is empty", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				CSVColumn: "sample",
			},
		}

		// act
		data, err := ExportBatch(e1, colMap)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, errors.New("param ExportColumnMap.DBColumn is required"))
	})

	t.Run("ordered selected fields will be exporter in the correct priority", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				CSVColumn: "age_custom",
				DBColumn:  "age",
			},
			{
				CSVColumn: "deleted_custom",
				DBColumn:  "deleted",
			},
			{
				CSVColumn: "some_name",
				DBColumn:  "some_name",
			},
		}

		// act
		res, err := ExportBatch(e1, colMap)

		// assert
		assert.Nil(t, err)

		assert.Equal(t, len(e1)+1, len(res))
		assert.Equal(t, []string{"age_custom", "deleted_custom", "some_name"}, res[0])

		assert.Equal(t, []string{"1", "1", n1}, res[1])
		assert.Equal(t, []string{"2", "0", n2}, res[2])
		assert.Equal(t, []string{"3", "0", n3}, res[3])
	})

	t.Run("success with slices and selected fields", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				CSVColumn: "name",
				DBColumn:  "some_name",
			},
			{
				CSVColumn: "age",
				DBColumn:  "age",
			},
		}

		// act
		res, err := ExportBatch(e1, colMap)

		// assert
		assert.Nil(t, err)

		assert.Equal(t, len(e1)+1, len(res))
		assert.Equal(t, []string{"name", "age"}, res[0])

		assert.Equal(t, []string{n1, "1"}, res[1])
		assert.Equal(t, []string{n2, "2"}, res[2])
		assert.Equal(t, []string{n3, "3"}, res[3])
	})

	t.Run("return header only with empty slices", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				DBColumn: "some_name",
			},
			{
				DBColumn: "age",
			},
		}
		expectedResp := [][]string{{"some_name", "age"}}

		// act
		res, err := ExportBatch([]database.Entity{}, colMap)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, res, expectedResp)
	})

	t.Run("success with boolean and float value - value types", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				CSVColumn: "some_name",
				DBColumn:  "some_name",
			},
			{
				DBColumn: "path",
			},
			{
				DBColumn: "age",
			},
			{
				DBColumn: "deleted",
			},
		}

		// act
		res, err := ExportBatch(e2, colMap)

		// assert
		assert.Nil(t, err)

		assert.Equal(t, len(e1)+1, len(res))
		assert.Equal(t, []string{"some_name", "path", "age", "deleted"}, res[0])

		assert.Equal(t, []string{n1_2, p1_2, "10.55", "1"}, res[1])
		assert.Equal(t, []string{n2_2, p2_2, "2.3333", "1"}, res[2])
		assert.Equal(t, []string{n3_2, p3_2, "3.444", "0"}, res[3])
	})

	t.Run("success with boolean and float value - pointer types", func(t *testing.T) {
		// arrange
		colMap := []ExportColumnMap{
			{
				CSVColumn: "path_custom",
				DBColumn:  "path",
			},
			{
				CSVColumn: "age",
				DBColumn:  "age",
			},
			{
				DBColumn: "deleted",
			},
		}

		// act
		res, err := ExportBatch(e1, colMap)

		// assert
		assert.Nil(t, err)

		assert.Equal(t, len(e1)+1, len(res))
		assert.Equal(t, []string{"path_custom", "age", "deleted"}, res[0])

		assert.Equal(t, []string{p1, "1", "1"}, res[1])
		assert.Equal(t, []string{p2, "2", "0"}, res[2])
		assert.Equal(t, []string{p3, "3", "0"}, res[3])
	})
}

func TestToCSV(t *testing.T) {
	t.Parallel()
	s := [][]string{
		{"title", "number", "content"},
		{"text", "1", `some tricky " text`},
		{"text2", "2", `some tricky "" text2`},
	}

	t.Run("escape double quotes", func(t *testing.T) {
		res := ToCSV(s)
		str := string(res)
		exp := `"title","number","content"` + "\n" +
			`"text","1","some tricky "" text"` + "\n" +

			`"text2","2","some tricky """" text2"` + "\n"
		assert.Equal(t, exp, str)
	})
}

func TestTransform(t *testing.T) {
	t.Parallel()

	t.Run("converts numeric value with decimal value to string", func(t *testing.T) {
		val := pgtype.Numeric{
			Int:    big.NewInt(int64(1060)),
			Exp:    -2,
			Status: pgtype.Present,
		}
		expectedStr := "10.6"

		res := transform(&val)
		str := string(res)
		assert.Equal(t, expectedStr, str)
	})
	t.Run("converts numeric value without decimal value to string", func(t *testing.T) {
		val := pgtype.Numeric{
			Int:    big.NewInt(int64(1000)),
			Exp:    -2,
			Status: pgtype.Present,
		}
		expectedStr := "10"

		res := transform(&val)
		str := string(res)
		assert.Equal(t, expectedStr, str)
	})
}

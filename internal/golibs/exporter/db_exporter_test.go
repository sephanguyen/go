package exporter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDBToCSV(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	t.Run("should return error if column map is empty", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()
		colMap := []ExportColumnMap{}
		numberOfRow := 10

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, errors.New("column map should not be empty"))
	})

	t.Run("should return error if DBColumn is empty", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()
		colMap := []ExportColumnMap{
			{
				CSVColumn: "sample",
			},
		}
		numberOfRow := 10

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, errors.New("param ExportColumnMap.DBColumn is required"))
	})

	t.Run("retrieve data failed", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		colMap := []ExportColumnMap{
			{
				DBColumn:  "some_name",
				CSVColumn: "name",
			},
			{
				DBColumn: "age",
			},
			{
				DBColumn:  "path",
				CSVColumn: "path",
			},
		}
		numberOfRow := 10
		ctx := context.Background()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, numberOfRow, 0)

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, data)
	})

	t.Run("should export data in correct order of fields", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()
		colMap := []ExportColumnMap{
			{
				DBColumn:  "some_name",
				CSVColumn: "name",
			},
			{
				DBColumn: "age",
			},
			{
				DBColumn:  "path",
				CSVColumn: "path",
			},
		}
		numberOfRow := 10
		sourceDataBytes := []byte(`"name","age","path"` + "\n" +
			`"Name 1","1","Path 1"` + "\n" +
			`"Name 2","2","Path 2"` + "\n")

		mockExportData(t, mockDB, numberOfRow, 0)

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, sourceDataBytes, data)
	})

	t.Run("should export data in correct title of csv data", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()
		colMap := []ExportColumnMap{
			{
				DBColumn:  "some_name",
				CSVColumn: "name_sample",
			},
			{
				DBColumn: "age",
			},
			{
				DBColumn:  "path",
				CSVColumn: "path_custom",
			},
		}
		numberOfRow := 10
		sourceDataBytes := []byte(`"name_sample","age","path_custom"` + "\n" +
			`"Name 1","1","Path 1"` + "\n" +
			`"Name 2","2","Path 2"` + "\n")
		mockExportData(t, mockDB, numberOfRow, 0)

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, sourceDataBytes, data)
	})
}

func mockExportData(t *testing.T, mockDB *testutil.MockDB, limit int, offset int) {
	rows := mockDB.Rows
	mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, limit, offset)
	rows.On("Next").Once().Return(true)
	rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Once().Run(func(args mock.Arguments) {
		reflect.ValueOf(args[0]).Elem().SetString("Name 1")
		reflect.ValueOf(args[1]).Elem().SetString("Path 1")
		reflect.ValueOf(args[2]).Elem().SetInt(1)
		reflect.ValueOf(args[3]).Elem().SetBool(false)
	}).Return(nil)
	rows.On("Next").Once().Return(true)
	rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Once().Run(func(args mock.Arguments) {
		reflect.ValueOf(args[0]).Elem().SetString("Name 2")
		reflect.ValueOf(args[1]).Elem().SetString("Path 2")
		reflect.ValueOf(args[2]).Elem().SetInt(2)
		reflect.ValueOf(args[3]).Elem().SetBool(false)
	}).Return(nil)
	rows.On("Next").Once().Return(false)
	rows.On("Close").Once().Return(nil)
}

func TestRetrieveAllData(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	t.Run("Fail case: Error when query data in database", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()

		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		data, err := RetrieveAllData(ctx, mockDB.DB, e)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, puddle.ErrClosedPool)
	})

	t.Run("Fail case: Error when scan entity data", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()

		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(puddle.ErrClosedPool)
		rows.On("Close").Once().Return(nil)

		data, err := RetrieveAllData(ctx, mockDB.DB, e)

		// assert
		assert.Nil(t, data)
		assert.Equal(t, err, fmt.Errorf("row.Scan: %w", puddle.ErrClosedPool))
	})

	t.Run("Happy case", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()

		rows := mockDB.Rows
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("Name 1")
			reflect.ValueOf(args[1]).Elem().SetString("Path 1")
			reflect.ValueOf(args[2]).Elem().SetInt(1)
			reflect.ValueOf(args[3]).Elem().SetBool(false)
		}).Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("Name 2")
			reflect.ValueOf(args[1]).Elem().SetString("Path 2")
			reflect.ValueOf(args[2]).Elem().SetInt(2)
			reflect.ValueOf(args[3]).Elem().SetBool(false)
		}).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)

		data, err := RetrieveAllData(ctx, mockDB.DB, e)

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, data)
	})

	t.Run("should export data in correct title of csv data", func(t *testing.T) {
		// arrange
		e := &TEntity{}
		ctx := context.Background()
		colMap := []ExportColumnMap{
			{
				DBColumn:  "some_name",
				CSVColumn: "name_sample",
			},
			{
				DBColumn: "age",
			},
			{
				DBColumn:  "path",
				CSVColumn: "path_custom",
			},
		}
		numberOfRow := 10
		sourceDataBytes := []byte(`"name_sample","age","path_custom"` + "\n" +
			`"Name 1","1","Path 1"` + "\n" +
			`"Name 2","2","Path 2"` + "\n")
		mockExportData(t, mockDB, numberOfRow, 0)

		// act
		data, err := DBToCSV(ctx, mockDB.DB, e, colMap, numberOfRow, 0)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, sourceDataBytes, data)
	})
}

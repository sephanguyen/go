package validators

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/stretchr/testify/assert"
)

type testEntity struct {
	Name    string
	Age     int
	School  string
	Deleted bool
}

func transform(s []string) (*testEntity, error) {
	name := s[0]
	age, err := strconv.ParseInt(s[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("can not parse %s to int", s[1])
	}

	deleted, err := strconv.ParseBool(s[2])
	if err != nil {
		return nil, fmt.Errorf("can not parse %s to bool: %s", s[1], err.Error())
	}
	return &testEntity{
		Name:    name,
		Age:     int(age),
		Deleted: deleted,
	}, nil
}

var colConfigs = []CSVColumn{
	{
		Column:   "name",
		Required: true,
	},
	{
		Column:   "age",
		Required: true,
	},
	{
		Column:   "deleted",
		Required: true,
	},
}
var config = CSVImportConfig[testEntity]{
	ColumnConfig: colConfigs,
	Transform:    transform,
}

func Test_ReadAndValidateCSV(t *testing.T) {
	t.Parallel()

	t.Run("should return error when transform function is nil", func(t *testing.T) {
		// arrange
		csvData := []byte("")
		config := CSVImportConfig[testEntity]{
			ColumnConfig: colConfigs,
		}
		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, data)
		assert.EqualError(t, err, "transform function must be defined")
	})

	t.Run("should return error when columns config is nil", func(t *testing.T) {
		// arrange
		csvData := []byte("")
		config := CSVImportConfig[testEntity]{
			Transform: transform,
		}
		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, data)
		assert.EqualError(t, err, "columns config must be defined")
	})

	t.Run("should return error when data is empty", func(t *testing.T) {
		// arrange
		csvData := []byte(`col1,col2,col3`)

		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, data)
		assert.EqualError(t, err, "no data in csv file")
	})

	t.Run("should return error when data has wrong number of columns", func(t *testing.T) {
		// arrange
		csvData := []byte(`name,age
		larry,11`)

		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, data)
		assert.EqualError(t, err, "wrong number of columns, expected 3, got 2")
	})

	t.Run("should return error when mandatory column is missing", func(t *testing.T) {
		// arrange
		csvData := []byte(`name,age,school
		larry,11,mana`)

		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, data)
		assert.EqualError(t, err, fmt.Sprintf("csv has invalid format, column number %d should be %s, got %s", 3, "deleted", "school"))
	})
}

func Test_ReadAndValidateCSV_ForLineValues(t *testing.T) {
	t.Parallel()
	// validate & parse line values
	t.Run("should return error when missing required value", func(t *testing.T) {
		// arrange
		csvData := []byte(`name,age,deleted
			bobby,10,0
			,11,1`)

		rowErr := &dto.UpsertError{
			RowNumber: 3,
			Error:     "column name is required",
		}

		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, 2, len(data))
		assert.EqualValues(t, rowErr, data[1].Error)
	})

	t.Run("should return parsing error from transformer", func(t *testing.T) {
		// arrange
		csvData := []byte(`name,age,deleted
			bobby,10z,0
			page,11,1`)

		rowErr := &dto.UpsertError{
			RowNumber: 2,
			Error:     "can not parse 10z to int",
		}

		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, 2, len(data))
		assert.EqualValues(t, rowErr, data[0].Error)
	})

	t.Run("should return correct value after parsing", func(t *testing.T) {
		// arrange
		csvData := []byte(`name,age,deleted
			bobby,10,0
			page,11,1`)

		e1 := &testEntity{
			Name:    "bobby",
			Age:     10,
			Deleted: false,
			School:  "",
		}
		e2 := &testEntity{
			Name:    "page",
			Age:     11,
			Deleted: true,
			School:  "",
		}
		// act
		data, err := ReadAndValidateCSV(csvData, config)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, 2, len(data))
		assert.EqualValues(t, e1, data[0].Value)
		assert.EqualValues(t, e2, data[1].Value)
	})
}

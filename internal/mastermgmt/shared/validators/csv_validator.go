package validators

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type CSVImportConfig[T any] struct {
	ColumnConfig []CSVColumn
	Transform    CSVLineTransformer[T]
}

type CSVColumn struct {
	Column   string
	Required bool
}

type CSVLineValue[T any] struct {
	Error *dto.UpsertError
	Value *T
}

func (c *CSVLineValue[T]) String() string {
	return fmt.Sprintf("%v", c.Value)
}

type CSVLineTransformer[T any] func([]string) (*T, error)

// ReadAndValidateCSV Read and validates the CSV file
// CSVColumn must be the same order with csv columns.
// CSV cell's spaces will be trimmed
func ReadAndValidateCSV[T any](payload []byte, config CSVImportConfig[T]) (data []*CSVLineValue[T], err error) {
	cols := config.ColumnConfig
	transform := config.Transform

	if transform == nil {
		return nil, fmt.Errorf("%s", "transform function must be defined")
	}
	if cols == nil {
		return nil, fmt.Errorf("%s", "columns config must be defined")
	}
	r := csv.NewReader(bytes.NewReader(payload))
	strLines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(strLines) < 2 {
		return nil, fmt.Errorf("%s", "no data in csv file")
	}

	// number of columns
	firstLine := strLines[0]

	if len(firstLine) != len(cols) {
		return nil, fmt.Errorf("wrong number of columns, expected %d, got %d", len(cols), len(firstLine))
	}

	// check mandatory columns
	for i, col := range firstLine {
		if strings.TrimSpace(cols[i].Column) != strings.TrimSpace(col) {
			return nil, fmt.Errorf("csv has invalid format, column number %d should be %s, got %s", (i + 1), cols[i].Column, col)
		}
	}

	// validate and parse
	for row, line := range strLines[1:] {
		lineValue := &CSVLineValue[T]{}
		for i, colConfig := range cols {
			// parse
			trimmedLine := sliceutils.Map(line, strings.TrimSpace)
			if colConfig.Required && strings.TrimSpace(trimmedLine[i]) == "" {
				lineValue.Error = &dto.UpsertError{
					RowNumber: int32(row + 2),
					Error:     fmt.Sprintf("column %s is required", colConfig.Column),
				}
			}

			// parse
			entity, err := transform(trimmedLine)
			if err != nil {
				lineValue.Error = &dto.UpsertError{
					RowNumber: int32(row + 2),
					Error:     err.Error(),
				}
			}
			lineValue.Value = entity
		}
		data = append(data, lineValue)
	}
	return data, nil
}

func GetErrorFromCSVValue[T any](c *CSVLineValue[T]) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fmt.Sprintf("Row Number: %d", c.Error.RowNumber),
		Description: c.Error.Error,
	}
}

func HasCSVErr[T any](c *CSVLineValue[T]) bool {
	return c.Error == nil
}

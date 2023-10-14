package exporter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"golang.org/x/exp/slices"
)

type ExportColumnMap struct {
	DBColumn  string // original name in DB
	CSVColumn string // re-mapped name, can be empty
}

// Get data from database and export to csv []byte
func DBToCSV(ctx context.Context, db database.QueryExecer, e database.Entity, columnMap []ExportColumnMap, limit int, offset int) (data []byte, err error) {
	csvCols, dbCols, err := validateColumnMap(columnMap)
	if err != nil {
		return nil, err
	}
	eData, err := retrieveData(ctx, db, e, limit, offset)
	if err != nil {
		return nil, err
	}

	strArr := make([][]string, len(eData)+1)

	// title column
	strArr[0] = csvCols

	for i, v := range eData {
		strLine := selectFields(v, dbCols)
		strArr[i+1] = strLine
	}

	return ToCSV(strArr), nil
}

func validateColumnMap(columnMap []ExportColumnMap) (csvCols []string, dbCols []string, err error) {
	if len(columnMap) == 0 {
		return nil, nil, errors.New("column map should not be empty")
	}
	csvCols = make([]string, len(columnMap))
	dbCols = make([]string, len(columnMap))
	for i, v := range columnMap {
		if v.DBColumn == "" {
			return nil, nil, errors.New("param ExportColumnMap.DBColumn is required")
		}
		if v.CSVColumn == "" {
			csvCols[i] = v.DBColumn
		} else {
			csvCols[i] = v.CSVColumn
		}

		dbCols[i] = v.DBColumn
	}
	return csvCols, dbCols, nil
}

// Select specified fields and order them
func selectFields(e database.Entity, dbCols []string) []string {
	allFields, values := e.FieldMap()

	singleLine := []string{}
	// order by selected fields
	for _, f := range dbCols {
		i := slices.Index(allFields, f)
		if i >= 0 {
			newStr := transform(values[i])
			singleLine = append(singleLine, newStr)
		}
	}
	return singleLine
}

// Get data from database
func retrieveData(ctx context.Context, db database.QueryExecer, e database.Entity, limit int, offset int) ([]database.Entity, error) {
	ctx, span := interceptors.StartSpan(ctx, "Golibs.DbExporter.retrieveData")
	defer span.End()

	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT $1 OFFSET $2", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(
		ctx,
		query,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []database.Entity
	for rows.Next() {
		// May be we have a better way to reflect the entity type from interface ?
		item := reflect.New(reflect.ValueOf(e).Elem().Type()).Interface().(database.Entity)

		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get all data from database without paging
func RetrieveAllData(ctx context.Context, db database.QueryExecer, e database.Entity) ([]database.Entity, error) {
	ctx, span := interceptors.StartSpan(ctx, "Golibs.DbExporter.RetrieveAllData")
	defer span.End()

	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(
		ctx,
		query,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []database.Entity
	for rows.Next() {
		// May be we have a better way to reflect the entity type from interface ?
		item := reflect.New(reflect.ValueOf(e).Elem().Type()).Interface().(database.Entity)

		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

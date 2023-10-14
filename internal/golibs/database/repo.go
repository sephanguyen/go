package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgconn"
)

// GeneratePlaceholders returns a string of "$1, $2, ..., $n".
func GeneratePlaceholders(n int) string {
	if n <= 0 {
		return ""
	}

	var builder strings.Builder
	sep := ", "
	for i := 1; i <= n; i++ {
		if i == n {
			sep = ""
		}
		builder.WriteString("$" + strconv.Itoa(i) + sep)
	}

	return builder.String()
}

// GeneratePlaceholdersWithFirstIndex returns a string of "$fi, $fi+1, ..., $fi+n-1".
func GeneratePlaceholdersWithFirstIndex(fi, n int) string {
	if fi < 1 {
		fi = 1
	}

	if n <= 0 {
		return ""
	}

	var builder strings.Builder
	sep := ", "
	for i := 0; i < n; i++ {
		if i == n-1 {
			sep = ""
		}
		builder.WriteString("$" + strconv.Itoa(fi+i) + sep)
	}

	return builder.String()
}

// InsertReturning entity using db Executor
func InsertReturning(
	ctx context.Context,
	e Entity,
	db QueryExecer,
	returnFieldName string,
	returnFieldValue interface{},
) error {
	fields := GetFieldNames(e)
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() +
		" (" + fieldNames + ") VALUES (" + placeHolders + ") RETURNING " + returnFieldName + ";"
	args := GetScanFields(e, fields)
	return db.QueryRow(ctx, stmt, args...).Scan(returnFieldValue)
}

// InsertReturningAndExcept entity using db Executor with option
// to exclude some fields (such as auto incremented id)
func InsertReturningAndExcept(
	ctx context.Context,
	e Entity,
	db QueryExecer,
	excludedFields []string,
	returnFieldName string,
	returnFieldValue interface{},
) error {
	fields := GetFieldNames(e)
	for _, ef := range excludedFields {
		fields = removeStrFromSlice(fields, ef)
	}
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() +
		" (" + fieldNames + ") VALUES (" + placeHolders + ") RETURNING " + returnFieldName + ";"
	args := GetScanFields(e, fields)
	return db.QueryRow(ctx, stmt, args...).Scan(returnFieldValue)
}

// Insert entity using db Executor
func Insert(
	ctx context.Context,
	e Entity,
	db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error),
) (pgconn.CommandTag, error) {
	fields := GetFieldNames(e)
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ");"
	args := GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

// InsertOnConflictDoNothing entity using db Executor
func InsertOnConflictDoNothing(
	ctx context.Context,
	e Entity,
	db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error),
) (pgconn.CommandTag, error) {
	fields := GetFieldNames(e)
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ") ON CONFLICT DO NOTHING;"
	args := GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

// InsertExcept insert entity using db Executor with option
// to exclude some fields (such as auto incremented id)
func InsertExcept(
	ctx context.Context,
	e Entity, excludedFields []string,
	db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error),
) (pgconn.CommandTag, error) {
	fields := GetFieldNames(e)
	for _, ef := range excludedFields {
		fields = removeStrFromSlice(fields, ef)
	}
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ");"
	args := GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

// InsertExceptOnConflictDoNothing insert entity using db Executor with option
// to exclude some fields (such as auto incremented id)
func InsertExceptOnConflictDoNothing(
	ctx context.Context,
	e Entity, excludedFields []string,
	db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error),
) (pgconn.CommandTag, error) {
	fields := GetFieldNames(e)
	for _, ef := range excludedFields {
		fields = removeStrFromSlice(fields, ef)
	}
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ") ON CONFLICT DO NOTHING;"
	args := GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

func removeStrFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// Insert entity using db Executor ignore conflict
func InsertIgnoreConflict(
	ctx context.Context,
	e Entity,
	db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error),
) (pgconn.CommandTag, error) {
	fields := GetFieldNames(e)
	fieldNames := strings.Join(fields, ",")
	placeHolders := GeneratePlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ") ON CONFLICT DO NOTHING;"
	args := GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

// TrimFieldEntity comprises of an entity E with a value N indicating the number of fields
// at the beginning of E that will be trimmed away when queried.
type TrimFieldEntity struct {
	E Entity
	N int
}

// TableName returns table name of entity.
func (te TrimFieldEntity) TableName() string {
	return te.E.TableName()
}

// FieldMap returns a trimmed slice of field names and values from entity. If N < 0,
// FieldMap returns all field names and values. If N is larger than the number of fields,
// it returns empty slices.
func (te TrimFieldEntity) FieldMap() ([]string, []interface{}) {
	fields, values := te.E.FieldMap()
	if te.N <= 0 {
		return fields, values
	}
	if te.N > len(fields) {
		return []string{}, []interface{}{}
	}
	return fields[te.N:], values[te.N:]
}

func generateUpdatePlaceholders(fields []string) string {
	var builder strings.Builder
	sep := ", "

	totalField := len(fields)
	for i, field := range fields {
		if i == totalField-1 {
			sep = ""
		}

		builder.WriteString(field + " = $" + strconv.Itoa(i+1) + sep)
	}

	return builder.String()
}

func GenerateUpdatePlaceholders(fields []string, firstIndex int) string {
	if firstIndex < 1 {
		firstIndex = 1
	}

	var builder strings.Builder
	sep := ", "

	totalField := len(fields)
	for i, field := range fields {
		if i == totalField-1 {
			sep = ""
		}

		builder.WriteString(field + " = $" + strconv.Itoa(i+firstIndex) + sep)
	}

	return builder.String()
}

func Update(ctx context.Context, e Entity, exec func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error), primaryField string) (pgconn.CommandTag, error) {
	var fields []string
	for _, field := range GetFieldNames(e) {
		if field != primaryField {
			fields = append(fields, field)
		}
	}
	placeHolders := generateUpdatePlaceholders(fields)

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE "+primaryField+" = $%d;", e.TableName(), placeHolders, len(fields)+1)

	args := GetScanFields(e, fields)
	primaryFieldValue := GetScanFields(e, []string{primaryField})[0]
	args = append(args, primaryFieldValue)

	return exec(ctx, stmt, args...)
}

func UpdateFields(ctx context.Context, e Entity, exec func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error), primaryField string, fields []string) (pgconn.CommandTag, error) {
	placeHolders := generateUpdatePlaceholders(fields)

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE "+primaryField+" = $%d;", e.TableName(), placeHolders, len(fields)+1)

	args := GetScanFields(e, fields)
	primaryFieldValue := GetScanFields(e, []string{primaryField})[0]
	args = append(args, primaryFieldValue)

	return exec(ctx, stmt, args...)
}

func UpdateFieldsForVersionNumber(ctx context.Context, e Entity, exec func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error), primaryField string, fields []string, versionNumber int32) (pgconn.CommandTag, error) {
	placeHolders := generateUpdatePlaceholders(fields)

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE "+primaryField+" = $%d AND version_number = %d;", e.TableName(), placeHolders, len(fields)+1, versionNumber)

	args := GetScanFields(e, fields)
	primaryFieldValue := GetScanFields(e, []string{primaryField})[0]
	args = append(args, primaryFieldValue)

	return exec(ctx, stmt, args...)
}

// FindColumn returns the index of targetColumn in columnNames slice or -1 if not found.
func FindColumn(columnNames []string, targetColumn string) int {
	for i, v := range columnNames {
		if v == targetColumn {
			return i
		}
	}
	return -1
}

// AddPagingQuery appends to the query string the LIMIT and OFFSET arguments.
// It returns the resulting query string and all the arguments in a slice.
func AddPagingQuery(query string, limit int32, page int32, args ...interface{}) (string, []interface{}) {
	if page > 0 {
		if limit == 0 {
			limit = 10
		}

		totalArgs := len(args)
		args = append(args, limit, limit*(page-1))
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", totalArgs+1, totalArgs+2)
	}
	return query, args
}

func compositeKeyPlaceHolder(sb *strings.Builder, counter *int, numKey int) {
	if numKey <= 0 {
		sb.WriteString("()")
		return
	}

	sb.WriteString("($" + strconv.Itoa(*counter))
	*counter = *counter + 1

	for i := 1; i < numKey; i++ {
		sb.WriteString(", $" + strconv.Itoa(*counter))
		*counter = *counter + 1
	}
	sb.WriteString(")")
}

// CompositeKeysPlaceHolders generate placeholders "($1, $2, $...), (...), ..." and arguments slice for composite keys.
func CompositeKeysPlaceHolders(n int, iterator func(i int) []interface{}) (string, []interface{}) {
	if n <= 0 {
		return "", nil
	}

	var sb strings.Builder
	counter := 1
	// for the first elevemt
	firstKeysBatch := iterator(0)
	args := make([]interface{}, 0, n*len(firstKeysBatch))

	args = append(args, firstKeysBatch...)
	compositeKeyPlaceHolder(&sb, &counter, len(firstKeysBatch))

	for idx := 1; idx < n; idx++ {
		keys := iterator(idx)
		args = append(args, keys...)
		sb.WriteString(", ")
		compositeKeyPlaceHolder(&sb, &counter, len(keys))
	}

	return sb.String(), args
}

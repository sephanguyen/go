package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
)

func generateInsertPlaceholders(n int) string {
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

type Entity interface {
	TableName() string
	FieldMap() map[string]pgtype.Value
}

func Insert(ctx context.Context, e database.Entity, db func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)) (pgconn.CommandTag, error) {
	fields := database.GetFieldNames(e)
	fieldNames := strings.Join(fields, ",")
	placeHolders := generateInsertPlaceholders(len(fields))
	stmt := "INSERT INTO " + e.TableName() + " (" + fieldNames + ") VALUES (" + placeHolders + ");"
	args := database.GetScanFields(e, fields)
	return db(ctx, stmt, args...)
}

func Update(ctx context.Context, e database.Entity, exec func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error), primaryField string) (pgconn.CommandTag, error) {
	var fields []string
	for _, field := range database.GetFieldNames(e) {
		if field != primaryField {
			fields = append(fields, field)
		}
	}
	placeHolders := generateUpdatePlaceholders(fields)

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE "+primaryField+" = $%d;", e.TableName(), placeHolders, len(fields)+1)

	args := database.GetScanFields(e, fields)
	primaryFieldValue := database.GetScanFields(e, []string{primaryField})[0]
	args = append(args, primaryFieldValue)

	return exec(ctx, stmt, args...)
}

package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/constants"
	golibs "github.com/manabie-com/backend/internal/golibs/database"
)

// GeneratePlaceHolderForBulkUpsert returns a string of "($1, $2, $3), ($4, $5 , $6), ..."
// nuOfItems: number of sets of values
// nuOfFields: number of values in each set
func GeneratePlaceHolderForBulkUpsert(nuOfItems int, nuOfField int) string {
	if nuOfField <= 0 || nuOfItems <= 0 {
		return ""
	}
	var builder strings.Builder
	count := 1
	for i := 1; i <= nuOfItems; i++ {
		builder.WriteString("(")
		for j := 1; j <= nuOfField; j++ {
			if j == nuOfField {
				builder.WriteString("$" + strconv.Itoa(count))
			} else {
				builder.WriteString("$" + strconv.Itoa(count) + ", ")
			}
			count++
		}
		if i == nuOfItems {
			builder.WriteString(")")
		} else {
			builder.WriteString("), ")
		}
	}

	return builder.String()
}

// BulkUpsert Upsert multiple value of entity T
func BulkUpsertAfterSplit[T golibs.Entity](ctx context.Context, db golibs.QueryExecer, prepareQuery string, items []T) error {
	fieldNames := golibs.GetFieldNames(items[0])
	placeHolders := GeneratePlaceHolderForBulkUpsert(len(items), len(fieldNames))
	query := fmt.Sprintf(prepareQuery, items[0].TableName(), strings.Join(fieldNames, ","), placeHolders)
	var scanFields []interface{}
	for _, v := range items {
		fields := golibs.GetScanFields(v, fieldNames)
		scanFields = append(scanFields, fields...)
	}
	ct, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no row affected")
	}
	return nil
}

// maxNumOfItem: the largest number of items possible so as not to exceed the number of params
func BulkUpsert[T golibs.Entity](ctx context.Context, db golibs.QueryExecer, prepareQuery string, items []T) error {
	if len(items) == 0 {
		return nil
	}
	fieldNames := golibs.GetFieldNames(items[0])
	maxNumOfItem := constants.LimitParamOfQuery / len(fieldNames)
	for i := 0; i < len(items); i += maxNumOfItem {
		if i+maxNumOfItem > len(items) {
			err := BulkUpsertAfterSplit(ctx, db, prepareQuery, items[i:])
			return err
		}
		err := BulkUpsertAfterSplit(ctx, db, prepareQuery, items[i:(i+maxNumOfItem)])
		if err != nil {
			return err
		}
	}
	return nil
}

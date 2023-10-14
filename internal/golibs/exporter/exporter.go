package exporter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/jackc/pgtype"
)

// Convert lines to csv file with escape character "
func ToCSV(lines [][]string) []byte {
	sb := strings.Builder{}
	for _, line := range lines {
		row := sliceutils.Map(line, getEscapedStr)
		//TODO: we need escape some special character when there are some complex data
		sb.WriteString(fmt.Sprintf("%s\n", strings.Join(row, ",")))
	}
	return []byte(sb.String())
}

// Get escaped tricky character like double quote
func getEscapedStr(s string) string {
	mustQuote := strings.ContainsAny(s, `"`)
	if mustQuote {
		s = strings.ReplaceAll(s, `"`, `""`)
	}

	return fmt.Sprintf(`"%s"`, s)
}

func ExportBatch(entities []database.Entity, columnMap []ExportColumnMap) ([][]string, error) {
	csvCols, dbCols, err := validateColumnMap(columnMap)
	if err != nil {
		return nil, err
	}

	str := make([][]string, len(entities)+1)

	// title column
	str[0] = csvCols

	for k, v := range entities {
		str[k+1] = selectFields(v, dbCols)
	}
	return str, nil
}

// transform the value with specific type's implementation.
// this method support both pointer types and value types.
// example: &age and age
func transform(v interface{}) string {
	val := reflect.ValueOf(v)
	newStr := ""

	switch val.Kind() {
	case reflect.Pointer:
		switch v.(type) {
		// pointer of boolean
		case *bool:
			newStr = transformBool(val.Elem().Bool())
		case *pgtype.Text:
			text := val.Interface().(*pgtype.Text)
			newStr = fmt.Sprint(text.String)
		case pgtype.Text:
			text := val.Interface().(pgtype.Text)
			newStr = fmt.Sprint(text.String)
		case *pgtype.Varchar:
			text := val.Interface().(*pgtype.Varchar)
			newStr = fmt.Sprint(text.String)
		case pgtype.Varchar:
			text := val.Interface().(pgtype.Varchar)
			newStr = fmt.Sprint(text.String)
		case *pgtype.Numeric:
			var floatVal float64
			text := val.Interface().(*pgtype.Numeric)
			_ = text.AssignTo(&floatVal)
			newStr = fmt.Sprint(floatVal)
		case pgtype.Numeric:
			var floatVal float64
			text := val.Interface().(pgtype.Numeric)
			_ = text.AssignTo(&floatVal)
			newStr = fmt.Sprint(floatVal)
		case *pgtype.Bool:
			b := val.Interface().(*pgtype.Bool)
			newStr = transformBool(b.Bool)
		case pgtype.Bool:
			b := val.Interface().(pgtype.Bool)
			newStr = transformBool(b.Bool)
		case pgtype.Date:
			d := val.Interface().(pgtype.Date)
			newStr = fmt.Sprint(d.Time.Format("2006-01-02"))
		case *pgtype.Date:
			d := val.Interface().(*pgtype.Date)
			newStr = fmt.Sprint(d.Time.Format("2006-01-02"))
		case pgtype.Timestamptz:
			d := val.Interface().(pgtype.Timestamptz)
			newStr = fmt.Sprint(d.Time.Format("2006-01-02 15:04:05"))
		case *pgtype.Timestamptz:
			d := val.Interface().(*pgtype.Timestamptz)
			newStr = fmt.Sprint(d.Time.Format("2006-01-02 15:04:05"))
		case pgtype.Int4:
			text := val.Interface().(pgtype.Int4)
			newStr = fmt.Sprint(text.Int)
		case *pgtype.Int4:
			text := val.Interface().(*pgtype.Int4)
			newStr = fmt.Sprint(text.Int)
		// TODO: Add more types if you need
		// other pointer types
		default:
			newStr = fmt.Sprint(val.Elem())
		}
	// value of boolean
	case reflect.Bool:
		newStr = transformBool(val.Bool())
	// other value types
	default:
		newStr = fmt.Sprint(val)
	}
	return newStr
}

func transformBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

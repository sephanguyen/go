package transformhelpers

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func StringToPgtypeText(str string) pgtype.Text {
	return database.Text(str)
}

func Int32ToPgtypeInt4(num int32) pgtype.Int4 {
	return database.Int4(num)
}

func ToPgtypeText(m protoreflect.Enum) pgtype.Text {
	return pgtype.Text{}
}

func BoolToPgtypeBool(b bool) pgtype.Bool {
	return database.Bool(b)
}

func Int32ToPgtypeInt4Array(arr []int32) pgtype.Int4Array {
	return database.Int4Array(arr)
}

func PgtypeTextToString(text pgtype.Text) string {
	return text.String
}

func PgtypeInt4ToInt32(num pgtype.Int4) int32 {
	return num.Int
}

func PgtypeTextTo(text pgtype.Text) protoreflect.Enum {
	return nil
}

func PgtypeBoolToBool(b pgtype.Bool) bool {
	return b.Bool
}

func PgtypeInt4ArrayToInt32(arr pgtype.Int4Array) []int32 {
	return database.FromInt4Array(arr)
}

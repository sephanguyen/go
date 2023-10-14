package database

import (
	"encoding/json"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Text converts a Go string to pgtype.Text.
func Text(v string) pgtype.Text {
	return pgtype.Text{String: v, Status: pgtype.Present}
}

// FromText returns the string data from t if it is present, nil otherwise.
func FromText(pgv pgtype.Text) *string {
	if pgv.Status != pgtype.Present {
		return nil
	}
	return &pgv.String
}

// Varchar converts a Go string to pgtype.Varchar. Varchar calls Text
// under the hood.
func Varchar(v string) pgtype.Varchar {
	return pgtype.Varchar(Text(v))
}

// FromVarchar returns the string data from t if it is present, nil otherwise.
// FromVarchar calls FromText under the hood.
func FromVarchar(pgv pgtype.Varchar) *string {
	return FromText(pgtype.Text(pgv))
}

// Int2 converts a Go int16 to pgtype.Int2.
func Int2(v int16) pgtype.Int2 {
	return pgtype.Int2{Int: v, Status: pgtype.Present}
}

// Int4 converts a Go int32 to pgtype.Int4.
func Int4(v int32) pgtype.Int4 {
	return pgtype.Int4{Int: v, Status: pgtype.Present}
}

// Int8 converts a Go int64 to pgtype.Int8.
func Int8(v int64) pgtype.Int8 {
	return pgtype.Int8{Int: v, Status: pgtype.Present}
}

// Float4 converts a Go float32 to pgtype.Float4.
func Float4(v float32) pgtype.Float4 {
	return pgtype.Float4{Float: v, Status: pgtype.Present}
}

// Bool converts a Go bool to pgtype.Bool.
func Bool(v bool) pgtype.Bool {
	return pgtype.Bool{Bool: v, Status: pgtype.Present}
}

// FromBool returns the bool data from b if it is present, nil otherwise.
func FromBool(pgv pgtype.Bool) *bool {
	if pgv.Status != pgtype.Present {
		return nil
	}
	return &pgv.Bool
}

// BoolArray converts a Go []bool to pgtype.BoolArray.
func BoolArray(v []bool) pgtype.BoolArray {
	if v == nil {
		return pgtype.BoolArray{Status: pgtype.Null}
	}
	if len(v) == 0 {
		return pgtype.BoolArray{Status: pgtype.Present}
	}
	elements := make([]pgtype.Bool, len(v))
	for idx := range v {
		elements[idx] = Bool(v[idx])
	}
	return pgtype.BoolArray{
		Elements:   elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

// FromBoolArray converts a Go pgtype.BoolArray to []bool.
func FromBoolArray(pgv pgtype.BoolArray) []bool {
	if pgv.Status != pgtype.Present {
		return nil
	}
	out := make([]bool, 0, len(pgv.Elements))
	for _, t := range pgv.Elements {
		out = append(out, t.Bool)
	}
	return out
}

// Int4Array converts a Go []int32 to pgtype.Int4Array.
func Int4Array(v []int32) pgtype.Int4Array {
	if v == nil {
		return pgtype.Int4Array{Status: pgtype.Null}
	}
	if len(v) == 0 {
		return pgtype.Int4Array{Status: pgtype.Present}
	}
	elements := make([]pgtype.Int4, len(v))
	for idx := range v {
		elements[idx] = Int4(v[idx])
	}
	return pgtype.Int4Array{
		Elements:   elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

func FromInt4Array(pgv pgtype.Int4Array) []int32 {
	if pgv.Status != pgtype.Present {
		return nil
	}
	out := make([]int32, 0, len(pgv.Elements))
	for _, t := range pgv.Elements {
		out = append(out, t.Int)
	}
	return out
}

func int4ArrayToIntArray[T any](pgv pgtype.Int4Array, getVal func(pgtype.Int4) T) []T {
	if pgv.Status != pgtype.Present {
		return nil
	}
	out := make([]T, 0, len(pgv.Elements))
	for _, t := range pgv.Elements {
		out = append(out, getVal(t))
	}

	return out
}

func getValueInt(t pgtype.Int4) int {
	return int(t.Int)
}

func getValueInt32(t pgtype.Int4) int32 {
	return t.Int
}

func Int4ArrayToIntArray(pgv pgtype.Int4Array) []int {
	return int4ArrayToIntArray(pgv, getValueInt)
}

func Int4ArrayToInt32Array(pgv pgtype.Int4Array) []int32 {
	return int4ArrayToIntArray(pgv, getValueInt32)
}

// Int8Array converts a Go []int64 to pgtype.Int8Array.
func Int8Array(v []int64) pgtype.Int8Array {
	if v == nil {
		return pgtype.Int8Array{Status: pgtype.Null}
	}
	if len(v) == 0 {
		return pgtype.Int8Array{Status: pgtype.Present}
	}
	elements := make([]pgtype.Int8, len(v))
	for idx := range v {
		elements[idx] = Int8(v[idx])
	}
	return pgtype.Int8Array{
		Elements:   elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

// TextArray converts a Go []string to pgtype.TextArray.
func TextArray(v []string) pgtype.TextArray {
	if v == nil {
		return pgtype.TextArray{Status: pgtype.Null}
	}
	if len(v) == 0 {
		return pgtype.TextArray{Status: pgtype.Present}
	}
	elements := make([]pgtype.Text, len(v))
	for i := range v {
		elements[i] = Text(v[i])
	}
	return pgtype.TextArray{
		Elements:   elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

// TextArrayVariadic is simply TextArray, but receives variadic strings instead.
func TextArrayVariadic(s ...string) pgtype.TextArray {
	return TextArray(s)
}

// FromTextArray returns the array data from ta if it is present, nil otherwise.
func FromTextArray(pgv pgtype.TextArray) []string {
	if pgv.Status != pgtype.Present {
		return nil
	}
	out := make([]string, 0, len(pgv.Elements))
	for _, t := range pgv.Elements {
		out = append(out, t.String)
	}
	return out
}

// MapFromTextArray converts TextArray to a lookup map.
func MapFromTextArray(ta pgtype.TextArray) map[pgtype.Text]bool {
	res := make(map[pgtype.Text]bool)
	for _, s := range ta.Elements {
		res[s] = true
	}
	return res
}

// JSONB converts a Go interface{} to pgtype.JSONB.
// This function ignores all error from json.Unmarshal. If b is a struct,
// ensure that it has valid json encoding.
func JSONB(v interface{}) pgtype.JSONB {
	j := pgtype.JSONB{}
	_ = j.Set(v)
	return j
}

// FromJSONB unmarshals jsonb data from pgv if it is present, does nothing otherwise.
// If out is nil or not a pointer, FromJSONB returns an InvalidUnmarshalError.
func FromJSONB(pgv pgtype.JSONB, out interface{}) error {
	if pgv.Status != pgtype.Present {
		return nil
	}
	return json.Unmarshal(pgv.Bytes, out)
}

// Timestamptz converts a time struct (e.g. time.Time) to pgtype.Timestamptz.
func Timestamptz(v time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: v, Status: pgtype.Present}
}

// TimestamptzNull converts a time struct (e.g. time.Time) to pgtype.Timestamptz and assign null value if v is zero value
func TimestamptzNull(v time.Time) pgtype.Timestamptz {
	if v.IsZero() {
		return pgtype.Timestamptz{Status: pgtype.Null}
	}
	return pgtype.Timestamptz{Time: v, Status: pgtype.Present}
}

// TimestamptzFromPb converts from google.golang.org/protobuf's timestamp to pgtype's timestamp.
// It returns a null-status pgtype.Timestamptz if input is nil.
func TimestamptzFromPb(v *timestamppb.Timestamp) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{Status: pgtype.Null}
	}
	return pgtype.Timestamptz{Time: v.AsTime(), Status: pgtype.Present}
}

func NewEmptyDate() pgtype.Date {
	return pgtype.Date{Status: pgtype.Null}
}

func DateFromPb(v *timestamppb.Timestamp) pgtype.Date {
	if v == nil {
		return NewEmptyDate()
	}
	return pgtype.Date{Time: v.AsTime(), Status: pgtype.Present}
}

// TimestamptzFromProto converts from github.com/gogo/protobuf's timestamp to pgtype's timestamp.
// It returns nil if input is nil.
func TimestamptzFromProto(v *types.Timestamp) (*pgtype.Timestamptz, error) {
	if v == nil {
		return nil, nil
	}
	t, err := types.TimestampFromProto(v)
	if err != nil {
		return nil, err
	}
	out := Timestamptz(t)
	return &out, nil
}

// FromTimestamptz returns the bool data from b if it is present, nil otherwise.
func FromTimestamptz(pgv pgtype.Timestamptz) *time.Time {
	if pgv.Status != pgtype.Present {
		return nil
	}
	return &pgv.Time
}

// Numeric converts a float32 to pgtype.Numeric. It uses pgtype.Numeric.Set but ignores
// any error from it. Ensure your input value is correct.
func Numeric(v float32) pgtype.Numeric {
	a := pgtype.Numeric{}
	_ = a.Set(v)
	return a
}

// Deprecated: use Timestamptz.
func TimeToPGTypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return Timestamptz(t)
}

func AppendText(ta pgtype.TextArray, elems ...pgtype.Text) pgtype.TextArray {
	ta.Elements = append(ta.Elements, elems...)
	return pgtype.TextArray{
		Elements:   ta.Elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(ta.Elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

func AppendTextArray(arr ...pgtype.TextArray) pgtype.TextArray {
	var textArr []pgtype.Text
	for _, i := range arr {
		textArr = append(textArr, i.Elements...)
	}

	var res pgtype.TextArray
	_ = res.Set(textArr)
	return res
}

// JSONBArray converts a Go []JSONB to pgtype.JSONBArray.
func JSONBArray(v []interface{}) pgtype.JSONBArray {
	if v == nil {
		return pgtype.JSONBArray{Status: pgtype.Null}
	}
	if len(v) == 0 {
		return pgtype.JSONBArray{Status: pgtype.Present}
	}
	elements := make([]pgtype.JSONB, len(v))
	for i := range v {
		elements[i] = JSONB(v[i])
	}
	return pgtype.JSONBArray{
		Elements:   elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

func AppendJSONB(ta pgtype.JSONBArray, elems ...pgtype.JSONB) pgtype.JSONBArray {
	ta.Elements = append(ta.Elements, elems...)
	return pgtype.JSONBArray{
		Elements:   ta.Elements,
		Dimensions: []pgtype.ArrayDimension{{Length: int32(len(ta.Elements)), LowerBound: 1}},
		Status:     pgtype.Present,
	}
}

// AppendJSONBProps
// Adds extra properties to a JSON object
// Only works for JSONB Object, not array
// It might sort properties by name.
func AppendJSONBProps(oldJSON pgtype.JSONB, values map[string]any) (newJSON pgtype.JSONB, err error) {
	if oldJSON.Status == pgtype.Null {
		err := newJSON.Set(values)
		if err != nil {
			return newJSON, err
		}
		return newJSON, nil
	}

	decoded := make(map[string]any)
	bytes, err := oldJSON.MarshalJSON()
	if err != nil {
		return newJSON, err
	}
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return newJSON, err
	}

	for key, val := range values {
		decoded[key] = val
	}

	newJSONBytes, err := json.Marshal(decoded)
	if err != nil {
		return newJSON, err
	}

	if err = newJSON.Set(newJSONBytes); err != nil {
		return newJSON, err
	}

	return newJSON, nil
}

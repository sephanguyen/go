package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestText(t *testing.T) {
	t.Parallel()
	testdata := []string{"", "a string"}
	for _, data := range testdata {
		actual := Text(data)
		assert.Exactly(t, data, actual.String)
		assert.Exactly(t, pgtype.Present, actual.Status)

		var expected pgtype.Text
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestFromText(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Text{
			String: "a string",
			Status: pgtype.Null,
		}
		assert.Nil(t, FromText(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Text{
			String: "a string",
			Status: pgtype.Present,
		}
		expected := "a string"
		assert.Exactly(t, &expected, FromText(data))
	})
}

func TestVarchar(t *testing.T) {
	t.Parallel()
	testdata := []string{"", "a string"}
	for _, data := range testdata {
		actual := Varchar(data)
		assert.Exactly(t, data, actual.String)
		assert.Exactly(t, pgtype.Present, actual.Status)

		var expected pgtype.Varchar
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestFromVarchar(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Varchar{
			String: "a string",
			Status: pgtype.Null,
		}
		assert.Nil(t, FromVarchar(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Varchar{
			String: "a string",
			Status: pgtype.Present,
		}
		expected := "a string"
		assert.Exactly(t, &expected, FromVarchar(data))
	})
}

func TestInt2(t *testing.T) {
	t.Parallel()
	var v int16 = 1<<15 - 1
	actual := Int2(v)
	assert.Exactly(t, v, actual.Int)
	assert.Exactly(t, pgtype.Present, actual.Status)

	var expected pgtype.Int2
	err := expected.Set(v)
	require.NoError(t, err)
	assert.Exactly(t, expected, actual)
}

func TestInt4(t *testing.T) {
	t.Parallel()
	var v int32 = 1<<31 - 1
	actual := Int4(v)
	assert.Exactly(t, v, actual.Int)
	assert.Exactly(t, pgtype.Present, actual.Status)

	var expected pgtype.Int4
	err := expected.Set(v)
	require.NoError(t, err)
	assert.Exactly(t, expected, actual)
}

func TestInt8(t *testing.T) {
	t.Parallel()
	var v int64 = 1<<63 - 1
	actual := Int8(v)
	assert.Exactly(t, v, actual.Int)
	assert.Exactly(t, pgtype.Present, actual.Status)

	var expected pgtype.Int8
	err := expected.Set(v)
	require.NoError(t, err)
	assert.Exactly(t, expected, actual)
}

func TestFloat4(t *testing.T) {
	t.Parallel()
	var v float32 = 12.34
	actual := Float4(v)
	assert.Exactly(t, v, actual.Float)
	assert.Exactly(t, pgtype.Present, actual.Status)

	var expected pgtype.Float4
	err := expected.Set(v)
	require.NoError(t, err)
	assert.Exactly(t, expected, actual)
}

func TestNumeric(t *testing.T) {
	t.Parallel()
	var v float32 = 12.3456
	var tmp float32
	actual := Numeric(v)
	actual.AssignTo(&tmp)
	assert.Exactly(t, v, tmp)
	assert.Exactly(t, pgtype.Present, actual.Status)

	var expected pgtype.Numeric
	err := expected.Set(v)
	require.NoError(t, err)
	assert.Exactly(t, expected, actual)
}

func TestBool(t *testing.T) {
	t.Parallel()
	testdata := []bool{false, true}
	for _, data := range testdata {
		actual := Bool(data)
		assert.Exactly(t, data, actual.Bool)
		assert.Exactly(t, pgtype.Present, actual.Status)

		var expected pgtype.Bool
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestFromBool(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Bool{
			Bool:   true,
			Status: pgtype.Null,
		}
		assert.Nil(t, FromBool(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Bool{
			Bool:   true,
			Status: pgtype.Present,
		}
		expected := true
		assert.Exactly(t, &expected, FromBool(data))
	})
}

func TestBoolArray(t *testing.T) {
	t.Parallel()
	testdata := [][]bool{
		nil,
		{},
		{true, false},
	}
	for _, data := range testdata {
		actual := BoolArray(data)
		for i := range data {
			assert.Exactly(t, data[i], actual.Elements[i].Bool)
			assert.Exactly(t, pgtype.Present, actual.Elements[i].Status)
		}

		var expected pgtype.BoolArray
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestFromBoolArray(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := BoolArray([]bool{true, false})
		data.Status = pgtype.Null
		assert.Nil(t, FromBoolArray(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := BoolArray([]bool{true, false})
		expected := []bool{true, false}
		assert.Exactly(t, expected, FromBoolArray(data))
	})
}

func TestInt4Array(t *testing.T) {
	t.Parallel()
	testdata := [][]int32{
		nil,
		{},
		{-4, -3, -2, -1, 0, 1, 2, 3, 4},
	}
	for _, data := range testdata {
		actual := Int4Array(data)
		for i := range data {
			assert.Exactly(t, data[i], actual.Elements[i].Int)
			assert.Exactly(t, pgtype.Present, actual.Elements[i].Status)
		}

		var expected pgtype.Int4Array
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestTextArray(t *testing.T) {
	t.Parallel()
	testdata := [][]string{
		nil,
		{},
		{"this", "is", "a", "list", "of", "strings"},
	}
	for _, data := range testdata {
		actual := TextArray(data)
		for i := range data {
			assert.Exactly(t, data[i], actual.Elements[i].String)
			assert.Exactly(t, pgtype.Present, actual.Elements[i].Status)
		}

		var expected pgtype.TextArray
		err := expected.Set(data)
		require.NoError(t, err)
		assert.Exactly(t, expected, actual)
	}
}

func TestTextArrayVariadic(t *testing.T) {
	t.Parallel()
	data := []string{"this", "is", "a", "list", "of", "strings"}
	actual := TextArrayVariadic(data...)
	for i := range data {
		assert.Exactly(t, data[i], actual.Elements[i].String)
		assert.Exactly(t, pgtype.Present, actual.Elements[i].Status)
	}
}

func TestFromTextArray(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := TextArray([]string{"abc", "123", "!@#"})
		data.Status = pgtype.Null
		assert.Nil(t, FromTextArray(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := TextArray([]string{"abc", "123", "!@#"})
		expected := []string{"abc", "123", "!@#"}
		assert.Exactly(t, expected, FromTextArray(data))
	})
}

func TestMapFromTextArray(t *testing.T) {
	t.Parallel()
	assert.Exactly(t,
		map[pgtype.Text]bool{Text("1"): true, Text("2"): true},
		MapFromTextArray(TextArrayVariadic("1", "2")),
	)
	assert.Exactly(t,
		map[pgtype.Text]bool{},
		MapFromTextArray(TextArray(nil)),
	)
}

type testTypeJSONB struct {
	Field1 string `json:"field_1"`
	Field2 string `json:"field_2"`
}

func TestJSONB(t *testing.T) {
	t.Parallel()
	data := []byte(`{"field_1": "value_1", "field_2": "value_2"}`)
	actual := JSONB(data)
	parsed := struct {
		Field1 string `json:"field_1"`
		Field2 string `json:"field_2"`
	}{}
	err := actual.AssignTo(&parsed)
	assert.NoError(t, err)
	assert.Exactly(t, parsed.Field1, "value_1")
	assert.Exactly(t, parsed.Field2, "value_2")
}

func TestFromJSONB(t *testing.T) {
	t.Parallel()
	data := testTypeJSONB{
		Field1: "abc",
		Field2: "123",
	}

	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		pgdata := JSONB(data)
		pgdata.Status = pgtype.Null

		var actual *testTypeJSONB
		err := FromJSONB(pgdata, actual)
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		pgdata := JSONB(data)
		actual := &testTypeJSONB{}
		err := FromJSONB(pgdata, actual)

		expected := &testTypeJSONB{
			Field1: "abc",
			Field2: "123",
		}
		assert.NoError(t, err)
		assert.Exactly(t, expected, actual)
	})

	t.Run("invalid argument", func(t *testing.T) {
		t.Parallel()
		pgdata := JSONB(data)
		var actual *testTypeJSONB
		err := FromJSONB(pgdata, actual)

		var expected *testTypeJSONB
		assert.EqualError(t, err, "json: Unmarshal(nil *database.testTypeJSONB)")
		assert.Exactly(t, expected, actual)
	})
}

// oldTimestamptz is the old Timestamptz function, kept for checking backward compatibility.
func oldTimestamptz(v interface{}) (*pgtype.Timestamptz, error) {
	ts := new(pgtype.Timestamptz)
	var err error

	switch data := v.(type) {
	case *types.Timestamp:
		if data == nil {
			return nil, nil
		}
		var t time.Time
		t, err = types.TimestampFromProto(data)
		if err != nil {
			return nil, err
		}
		err = ts.Set(t)
	case time.Time:
		err = ts.Set(data)
	case *timestamppb.Timestamp:
		if data == nil {
			err = ts.Set(nil)
		} else {
			err = ts.Set(data.AsTime())
		}
	default:
		err = fmt.Errorf("unsupported value %v of type %T", data, v)
	}
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func TestTimestamptz(t *testing.T) {
	t.Parallel()
	t.Run("normal case", func(t *testing.T) {
		t.Parallel()
		input := time.Now()
		output := Timestamptz(input)
		assert.True(t, output.Time.Equal(input))
		assert.Exactly(t, pgtype.Present, output.Status)
		assert.Exactly(t, pgtype.None, output.InfinityModifier)

		// Compare with old version
		oldOutput, err := oldTimestamptz(input)
		require.NoError(t, err)
		assert.Exactly(t, *oldOutput, output)
	})
}

func TestTimestamptzNull(t *testing.T) {
	t.Parallel()
	t.Run("normal case", func(t *testing.T) {
		t.Parallel()
		input := time.Now()
		output := TimestamptzNull(input)
		assert.True(t, output.Time.Equal(input))
		assert.Exactly(t, pgtype.Present, output.Status)
		assert.Exactly(t, pgtype.None, output.InfinityModifier)

		// Compare with old version
		oldOutput, err := oldTimestamptz(input)
		require.NoError(t, err)
		assert.Exactly(t, *oldOutput, output)
	})
}

func TestTimestamptzFromPb(t *testing.T) {
	t.Parallel()
	t.Run("normal case", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		input := timestamppb.New(now)
		output := TimestamptzFromPb(input)
		assert.True(t, output.Time.Equal(now))
		assert.Exactly(t, pgtype.Present, output.Status)
		assert.Exactly(t, pgtype.None, output.InfinityModifier)

		// Compare with old version
		oldOutput, err := oldTimestamptz(input)
		require.NoError(t, err)
		assert.Exactly(t, *oldOutput, output)
	})

	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		var nilInput *timestamppb.Timestamp
		output := TimestamptzFromPb(nilInput)
		assert.True(t, output.Time.IsZero())
		assert.Exactly(t, pgtype.Null, output.Status)
		assert.Exactly(t, pgtype.None, output.InfinityModifier)

		// Compare with old version
		oldOutput, err := oldTimestamptz(nilInput)
		require.NoError(t, err)
		assert.Exactly(t, *oldOutput, output)
	})
}

func TestTimestamptzFromProto(t *testing.T) {
	t.Parallel()
	t.Run("normal case", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		input, err := types.TimestampProto(now)
		require.NoError(t, err)
		output, err := TimestamptzFromProto(input)
		require.NoError(t, err)
		assert.True(t, output.Time.Equal(now))
		assert.Exactly(t, pgtype.Present, output.Status)
		assert.Exactly(t, pgtype.None, output.InfinityModifier)

		// Compare with old version
		oldOutput, err := oldTimestamptz(input)
		require.NoError(t, err)
		assert.Exactly(t, oldOutput, output)
	})

	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		var nilInput *types.Timestamp
		output, err := TimestamptzFromProto(nilInput)
		require.NoError(t, err)
		assert.Nil(t, output)

		// Compare with old version
		oldOutput, err := oldTimestamptz(nilInput)
		require.NoError(t, err)
		assert.Exactly(t, oldOutput, output)
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		invalidInput := &types.Timestamp{Seconds: -1, Nanos: -2}
		actual, err := TimestamptzFromProto(invalidInput)
		assert.EqualError(t, err, "timestamp: &types.Timestamp{Seconds: -1,\nNanos: -2,\n}: nanos not in range [0, 1e9)")
		assert.Nil(t, actual)

		// Compare with old version
		oldOutput, err := oldTimestamptz(invalidInput)
		assert.EqualError(t, err, "timestamp: &types.Timestamp{Seconds: -1,\nNanos: -2,\n}: nanos not in range [0, 1e9)")
		assert.Nil(t, oldOutput)
	})
}

func TestFromTimestamptz(t *testing.T) {
	t.Parallel()
	t.Run("nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Timestamptz{
			Time:   time.Date(1234, 5, 6, 7, 8, 9, 10, time.Local),
			Status: pgtype.Null,
		}
		assert.Nil(t, FromTimestamptz(data))
	})

	t.Run("non-nil case", func(t *testing.T) {
		t.Parallel()
		data := pgtype.Timestamptz{
			Time:   time.Date(1234, 5, 6, 7, 8, 9, 10, time.Local),
			Status: pgtype.Present,
		}
		expected := time.Date(1234, 5, 6, 7, 8, 9, 10, time.Local)
		assert.Exactly(t, &expected, FromTimestamptz(data))
	})
}

func TestAppendText(t *testing.T) {
	t.Parallel()
	assert.Exactly(t,
		TextArrayVariadic("1", "2", "3", "4"),
		AppendText(TextArrayVariadic("1", "2"), Text("3"), Text("4")),
	)
	assert.Exactly(t,
		TextArrayVariadic("1", "2", "3", "4"),
		AppendText(pgtype.TextArray{}, Text("1"), Text("2"), Text("3"), Text("4")),
	)
}

func TestAppendJSONBProps(t *testing.T) {
	t.Parallel()

	t.Run("old JSONB is null, then take all new values to the new JSONB", func(t *testing.T) {
		// arrange
		oldJSON := pgtype.JSONB{}
		oldJSON.Set(nil)
		newRawJSON := json.RawMessage(`{"prop_a":"value_a","prop_b":1}`)
		newJSONB := pgtype.JSONB{Bytes: newRawJSON, Status: pgtype.Present}
		values := map[string]any{
			"prop_a": "value_a",
			"prop_b": 1,
		}

		// act
		actual, err := AppendJSONBProps(oldJSON, values)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, newJSONB, actual)
	})

	t.Run("new values is null or empty, then return the old JSONB", func(t *testing.T) {
		// arrange
		jsonRaw := json.RawMessage(`{"prop_a":"value_a","prop_b":1}`)
		oldJSON := pgtype.JSONB{Bytes: jsonRaw, Status: pgtype.Present}
		values := map[string]any{}
		var nilMap map[string]any

		// act
		actual, err := AppendJSONBProps(oldJSON, values)
		actual2, err2 := AppendJSONBProps(oldJSON, nilMap)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, oldJSON, actual)

		assert.Nil(t, err2)
		assert.Equal(t, oldJSON, actual2)
	})

	t.Run("old JSONB and new values is not null, append new values (overwrite)", func(t *testing.T) {
		// arrange
		oldJSONRaw := json.RawMessage(`{"a":"value_a","b":1,"c":2}`)
		oldJSON := pgtype.JSONB{Bytes: oldJSONRaw, Status: pgtype.Present}
		newJSONRaw := json.RawMessage(`{"a":"value_updated","b":1,"c":2,"d":"new_value","e":7.35}`)
		newJSON := pgtype.JSONB{Bytes: newJSONRaw, Status: pgtype.Present}
		values := map[string]any{
			"a": "value_updated",
			"d": "new_value",
			"e": 7.35,
			"c": 2,
		}

		// act
		actual, err := AppendJSONBProps(oldJSON, values)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, newJSON, actual)
	})
}

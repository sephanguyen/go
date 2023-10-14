package golibs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInArrayString(t *testing.T) {
	t.Parallel()
	stringArr := []string{"an", "array", "string"}
	t.Run("string in array", func(t *testing.T) {
		t.Parallel()
		result := InArrayString("array", stringArr)
		assert.Exactly(t, true, result)
	})

	t.Run("string not in array", func(t *testing.T) {
		t.Parallel()
		result := InArrayString("not in array", stringArr)
		assert.Exactly(t, false, result)
	})
}

func TestEqualStringArray(t *testing.T) {
	t.Parallel()
	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		result := EqualStringArray([]string{"a", "b"}, []string{"a", "b"})
		assert.Exactly(t, true, result)
	})

	t.Run("not equal", func(t *testing.T) {
		t.Parallel()
		result := EqualStringArray([]string{"a", "b"}, []string{"a", "c"})
		assert.Exactly(t, false, result)
	})
}

func TestUniq(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result := Uniq([]string{"a", "string", "string", "array", "array", "array"})
		assert.ElementsMatch(t, []string{"a", "string", "array"}, result)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result := Uniq([]string{"a", "unique", "string", "array"})
		assert.ElementsMatch(t, []string{"a", "unique", "string", "array"}, result)
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		result := Uniq(nil)
		assert.ElementsMatch(t, nil, result)
	})
}

func TestReplace(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result := Replace(
			[]string{"this", "is", "the", "original", "slice"},
			[]string{"this", "original"},
			[]string{"that", "new"},
		)
		assert.Exactly(t, []string{"that", "is", "the", "new", "slice"}, result)
	})

	t.Run("does nothing due to input length mismatch", func(t *testing.T) {
		t.Parallel()
		result := Replace(
			[]string{"this", "is", "the", "original", "slice"},
			[]string{"this", "slice"},
			[]string{"that"},
		)
		assert.Exactly(t, []string{"this", "is", "the", "original", "slice"}, result)
	})
}

func TestCompare(t *testing.T) {
	t.Parallel()
	arr1 := []string{"a", "b", "c", "d"}
	arr2 := []string{"f", "e", "d", "b"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		intersect, added, removed := Compare(arr1, arr2)
		assert.ElementsMatch(t, []string{"b", "d"}, intersect)
		assert.ElementsMatch(t, []string{"e", "f"}, added)
		assert.ElementsMatch(t, []string{"a", "c"}, removed)
	})

	t.Run("return nil", func(t *testing.T) {
		t.Parallel()
		intersect, added, removed := Compare(nil, arr2)
		assert.Nil(t, intersect)
		assert.Nil(t, added)
		assert.Nil(t, removed)

		intersect, added, removed = Compare(arr1, nil)
		assert.Nil(t, intersect)
		assert.Nil(t, added)
		assert.Nil(t, removed)
	})
}

func TestToArrayStringPostgres(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result := ToArrayStringPostgres([]int64{-1, 0, 1, 2, 3, 4, 5})
		assert.Exactly(t, "{-1,0,1,2,3,4,5}", result)
	})
}

func TestGetBrightcoveVideoIDFromURL(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		videoURL string
		videoID  string
		hasError bool
	}{
		{
			name:     "get video id from url",
			videoURL: "https://brightcove.com/account/123/video?videoId=abcd1234",
			videoID:  "abcd1234",
			hasError: false,
		},
		{
			name:     "get video id from invalid query param",
			videoURL: "https://brightcove.com/account/123/video?videoID=abcd1234",
			hasError: true,
		},
		{
			name:     "get video id from url has empty value of query param",
			videoURL: "https://brightcove.com/account/123/video?videoId=",
			hasError: true,
		},
		{
			name:     "get video id from url has no query param",
			videoURL: "https://brightcove.com/account/123/video",
			hasError: true,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actualID, err := GetBrightcoveVideoIDFromURL(tc.videoURL)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.videoID, actualID)
			}
		})
	}
}

func TestToArrayStringFromArrayInt64(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result := ToArrayStringFromArrayInt64([]int64{-1, 0, 1, 2, 3, 4, 5})
		assert.Exactly(t, []string{"-1", "0", "1", "2", "3", "4", "5"}, result)
	})

	t.Run("input nil", func(t *testing.T) {
		t.Parallel()
		result := ToArrayStringFromArrayInt64(nil)
		assert.Exactly(t, []string{}, result)
		assert.Exactly(t, 0, len(result))
	})

	t.Run("input not nil but empty", func(t *testing.T) {
		t.Parallel()
		result := ToArrayStringFromArrayInt64([]int64{})
		assert.Exactly(t, []string{}, result)
		assert.Exactly(t, 0, len(result))
	})
}

func TestGetContentType(t *testing.T) {
	t.Parallel()
	t.Run("success audio.mp3", func(t *testing.T) {
		t.Parallel()
		result := GetContentType("audio.mp3")
		assert.Equal(t, "audio/mpeg", result)
	})

	t.Run("success application.pdf", func(t *testing.T) {
		t.Parallel()
		result := GetContentType("application.pdf")
		assert.Equal(t, "application/pdf", result)
	})

	t.Run("success application.pdf", func(t *testing.T) {
		t.Parallel()
		result := GetContentType("bucket/application.pdf")
		assert.Equal(t, "application/pdf", result)
	})

	t.Run("success application.pdf", func(t *testing.T) {
		t.Parallel()
		result := GetContentType("/application.pdf")
		assert.Equal(t, "application/pdf", result)
	})

	t.Run("return default content type", func(t *testing.T) {
		t.Parallel()
		result := GetContentType("applicationpdf")
		assert.Equal(t, "application/octet-stream", result)
	})
}

func TestStack(t *testing.T) {
	t.Parallel()
	t.Run("push stack successfully", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		s.Push(2)
		assert.Equal(t, 2, (s.Elements[0]).(int))
		s.Push("example text")
		assert.Equal(t, "example text", (s.Elements[1]).(string))
	})

	t.Run("pop stack successfully", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		s.Push(2)
		res, err := s.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 2, res.(int))

		s.Push("example text")
		res, err = s.Pop()
		assert.NoError(t, err)
		assert.Equal(t, "example text", res.(string))
	})

	t.Run("pop empty stack ", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		s.Push(2)
		res, err := s.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 2, res.(int))
		assert.True(t, s.IsEmpty())

		res, err = s.Pop()
		assert.EqualError(t, err, "empty stack")
	})

	t.Run("peek stack successfully", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		s.Push(2)
		res, err := s.Peek()
		assert.NoError(t, err)
		assert.Equal(t, 2, res.(int))
		assert.False(t, s.IsEmpty())
	})

	t.Run("peek empty stack", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		_, err := s.Peek()
		assert.EqualError(t, err, "empty stack")
	})

	t.Run("peek multi stack successfully", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		listExpected := []interface{}{2, 3, "example text"}
		for _, expected := range listExpected {
			s.Push(expected)
		}
		res, err := s.PeekMulti(len(listExpected))
		assert.NoError(t, err)

		assert.False(t, s.IsEmpty())
		for i, expected := range listExpected {
			s.Push(expected)
			assert.EqualValues(t, expected, res[i])
		}
	})

	t.Run("peek multi stack with not enough items", func(t *testing.T) {
		s := Stack{Elements: []interface{}{}}
		assert.True(t, s.IsEmpty())

		listExpected := []interface{}{2, 3, "example text"}
		for _, expected := range listExpected {
			s.Push(expected)
		}
		_, err := s.PeekMulti(len(listExpected) + 1)
		assert.EqualError(t, err, "not enough items in stack")
	})
}

func TestTimestamppbToTime(t *testing.T) {
	tcs := []struct {
		input    *timestamppb.Timestamp
		expected time.Time
	}{
		{
			input:    nil,
			expected: time.Time{},
		},
	}

	for _, tc := range tcs {
		acctual := TimestamppbToTime(tc.input)
		assert.Equal(t, tc.expected, acctual)
	}
}

func TestAll(t *testing.T) {
	tcs := []struct {
		input    interface{}
		expected bool
	}{
		{
			input:    []int{1, 2},
			expected: true,
		},
		{
			input:    []int{0, 1},
			expected: false,
		},
		{
			input:    []int{0, 0},
			expected: false,
		},
		{
			input:    ([]int)(nil),
			expected: false,
		},
		{
			input:    []string{".", "."},
			expected: true,
		},
		{
			input:    []string{"", "."},
			expected: false,
		},
		{
			input:    [][]string{{""}},
			expected: true,
		},
		{
			input:    []interface{}{1, "", 2},
			expected: false,
		},
		{
			input:    []interface{}{[]interface{}{}, "", 2},
			expected: false,
		},
		{
			input:    []interface{}{3, ".", 2},
			expected: true,
		},
		{
			input:    []interface{}{".", 0},
			expected: false,
		},
		{
			input:    []string{"", ""},
			expected: false,
		},
		{
			input:    ([]string)(nil),
			expected: false,
		},

		{
			input:    2,
			expected: false,
		},

		{
			input:    "",
			expected: false,
		},
	}

	for index, tc := range tcs {
		fmt.Println("--", index)
		acctual := All(tc.input)
		assert.Equal(t, tc.expected, acctual)
	}
}

func TestStringSliceToMap(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		data := []string{"A", "B", "123", "B"}
		res := StringSliceToMap(data)
		assert.Len(t, res, 3)
		for _, d := range data {
			if _, ok := res[d]; !ok {
				assert.Fail(t, fmt.Sprintf("string %s is not in result", d))
			}
		}
	})
}

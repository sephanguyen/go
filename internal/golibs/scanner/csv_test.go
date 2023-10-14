package scanner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCSVScanner(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected [][]string
	}{
		{
			name: "scanner with all data",
			data: `name,course_id,location_id
			class-1,course-1,location-1
			class-2,course-2,location-2`,
			expected: [][]string{
				{"class-1", "course-1", "location-1"},
				{"class-2", "course-2", "location-2"},
			},
		},
		{
			name: "scanner with wrong name colum",
			data: `wrong-name,course_id,location_id
			class-1,course-1,location-1
			class-2,course-2,location-2`,
			expected: [][]string{
				{"", "course-1", "location-1"},
				{"", "course-2", "location-2"},
			},
		},
		{
			name:     "scanner with empty data",
			data:     `name,course_id,location_id`,
			expected: [][]string{},
		},
		{
			name: "scanner with empty head and data",
			data: "",
		},
	}

	for _, tc := range testCases {
		scanner := NewCSVScanner(strings.NewReader(tc.data))
		if len(tc.data) == 0 {
			require.Empty(t, scanner)
			continue
		}
		index := 0
		for scanner.Scan() {
			require.Equal(t, tc.expected[index][0], scanner.Text("name"))
			require.Equal(t, tc.expected[index][1], scanner.Text("course_id"))
			require.Equal(t, tc.expected[index][2], scanner.Text("location_id"))
			index += 1
		}
		require.Equal(t, index, len(tc.expected))
	}
}

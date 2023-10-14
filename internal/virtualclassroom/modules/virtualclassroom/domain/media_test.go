package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMedia_isValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name    string
		media   *Media
		isValid bool
	}{
		{
			name: "full fields",
			media: &Media{
				ID:        "test-id",
				Name:      "test-name",
				Resource:  "test-resource",
				Type:      MediaTypeImage,
				CreatedAt: now,
				Comments:  []Comment{{Comment: "test"}},
				ConvertedImages: []ConvertedImage{
					{Width: 1000, Height: 100},
					{Height: 1000},
					{ImageURL: "test-url"},
				},
			},
			isValid: true,
		},
		{
			name: "missing id",
			media: &Media{
				ID:        "",
				Name:      "test-name",
				Resource:  "",
				Type:      MediaTypeImage,
				CreatedAt: now,
				Comments:  []Comment{{Comment: "test"}},
				ConvertedImages: []ConvertedImage{
					{Width: 1000, Height: 100},
					{Height: 1000},
					{ImageURL: "test-url"},
				},
			},
			isValid: false,
		},
		{
			name: "missing converted image",
			media: &Media{
				ID:              "test-id",
				Name:            "test-name",
				Resource:        "",
				Type:            MediaTypeImage,
				CreatedAt:       now,
				Comments:        []Comment{{Comment: "test"}},
				ConvertedImages: nil,
			},
			isValid: false,
		},
		{
			name: "missing resource",
			media: &Media{
				ID:        "test-id",
				Name:      "test-name",
				Resource:  "",
				Type:      MediaTypeImage,
				CreatedAt: now,
				Comments:  []Comment{{Comment: "test"}},
				ConvertedImages: []ConvertedImage{
					{Width: 1000, Height: 100},
					{Height: 1000},
					{ImageURL: "test-url"},
				},
			},
			isValid: false,
		},
		{
			name: "missing type",
			media: &Media{
				ID:        "test-id",
				Name:      "test-name",
				Resource:  "test-resource",
				Type:      "",
				CreatedAt: now,
				Comments:  []Comment{{Comment: "test"}},
				ConvertedImages: []ConvertedImage{
					{Width: 1000, Height: 100},
					{Height: 1000},
					{ImageURL: "test-url"},
				},
			},
			isValid: false,
		},
		{
			name: "missing name",
			media: &Media{
				ID:        "test-id",
				Name:      "",
				Resource:  "test-resource",
				Type:      "test-type",
				CreatedAt: now,
				Comments:  []Comment{{Comment: "test"}},
				ConvertedImages: []ConvertedImage{
					{Width: 1000, Height: 100},
					{Height: 1000},
					{ImageURL: "test-url"},
				},
			},
			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.media.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

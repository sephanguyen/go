package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVideoPresentMaterialState_isValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name     string
		material *VideoPresentMaterialState
		isValid  bool
	}{
		{
			name: "full fields",
			material: &VideoPresentMaterialState{
				Material: &VideoMaterial{
					ID:      "id",
					Name:    "video",
					VideoID: "video-id",
				},
				UpdatedAt: now,
				VideoState: &VideoState{
					CurrentTime: Duration(2 * time.Minute),
					PlayerState: PlayerStatePlaying,
				},
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			material: &VideoPresentMaterialState{
				Material: &VideoMaterial{
					ID:      "id",
					VideoID: "video-id",
				},
				VideoState: &VideoState{
					PlayerState: PlayerStatePlaying,
				},
			},
			isValid: true,
		},
		{
			name: "miss material field",
			material: &VideoPresentMaterialState{
				UpdatedAt: now,
				VideoState: &VideoState{
					CurrentTime: Duration(2 * time.Minute),
					PlayerState: PlayerStatePlaying,
				},
			},
			isValid: false,
		},
		{
			name: "miss video state field",
			material: &VideoPresentMaterialState{
				Material: &VideoMaterial{
					ID:      "id",
					Name:    "video",
					VideoID: "video-id",
				},
				UpdatedAt: now,
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.material.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPDFPresentMaterialState_isValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name     string
		material *PDFPresentMaterialState
		isValid  bool
	}{
		{
			name: "full fields",
			material: &PDFPresentMaterialState{
				Material: &PDFMaterial{
					ID:   "id",
					Name: "video",
					URL:  "https://example.com/random-path/name.pdf",
					ConvertedImageURL: &ConvertedImage{
						Width:    2,
						Height:   3,
						ImageURL: "https://example.com/random-path/name.png",
					},
				},
				UpdatedAt: now,
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			material: &PDFPresentMaterialState{
				Material: &PDFMaterial{
					ID:  "id",
					URL: "https://example.com/random-path/name.pdf",
				},
			},
			isValid: true,
		},
		{
			name: "miss material field",
			material: &PDFPresentMaterialState{
				UpdatedAt: now,
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.material.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

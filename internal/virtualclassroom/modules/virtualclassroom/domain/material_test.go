package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoMaterial_isValid(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		material VideoMaterial
		isValid  bool
	}{
		{
			name: "full fields",
			material: VideoMaterial{
				ID:      "id",
				Name:    "video",
				VideoID: "video-id",
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			material: VideoMaterial{
				ID:      "id",
				VideoID: "video-id",
			},
			isValid: true,
		},
		{
			name: "miss id field",
			material: VideoMaterial{
				Name:    "video",
				VideoID: "video-id",
			},
			isValid: false,
		},
		{
			name: "miss video id field",
			material: VideoMaterial{
				ID:   "id",
				Name: "video",
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.material.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPDFMaterial_isValid(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		material PDFMaterial
		isValid  bool
	}{
		{
			name: "full fields",
			material: PDFMaterial{
				ID:   "id",
				Name: "pdf",
				URL:  "https://example.com/random-path/name.pdf",
				ConvertedImageURL: &ConvertedImage{
					Width:    2,
					Height:   3,
					ImageURL: "https://example.com/random-path/name.png",
				},
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			material: PDFMaterial{
				ID:  "id",
				URL: "https://example.com/random-path/name.pdf",
			},
			isValid: true,
		},
		{
			name: "miss id field",
			material: PDFMaterial{
				Name: "pdf",
				URL:  "https://example.com/random-path/name.pdf",
				ConvertedImageURL: &ConvertedImage{
					Width:    2,
					Height:   3,
					ImageURL: "https://example.com/random-path/name.png",
				},
			},
			isValid: false,
		},
		{
			name: "miss url field",
			material: PDFMaterial{
				ID:   "id",
				Name: "pdf",
				ConvertedImageURL: &ConvertedImage{
					Width:    2,
					Height:   3,
					ImageURL: "https://example.com/random-path/name.png",
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.material.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMaterials_GetMaterialElement(t *testing.T) {
	materials := Materials{
		&PDFMaterial{},
		&VideoMaterial{},
	}

	// test GetVideoMaterialElement func
	video := materials.GetVideoMaterialElement(0)
	assert.Nil(t, video)
	video = materials.GetVideoMaterialElement(1)
	assert.NotNil(t, video)
	video = materials.GetVideoMaterialElement(len(materials))
	assert.Nil(t, video)

	// test GetPDFMaterialElement func
	pdf := materials.GetPDFMaterialElement(0)
	assert.NotNil(t, pdf)
	pdf = materials.GetPDFMaterialElement(1)
	assert.Nil(t, pdf)
	pdf = materials.GetPDFMaterialElement(len(materials))
	assert.Nil(t, pdf)
}

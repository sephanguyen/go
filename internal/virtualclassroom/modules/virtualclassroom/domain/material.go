package domain

import (
	"fmt"
)

type Materials []material

func (m Materials) isValid() error {
	for _, e := range m {
		if err := e.isValid(); err != nil {
			return err
		}
	}
	return nil
}

func (m Materials) GetVideoMaterialElement(i int) *VideoMaterial {
	if i >= len(m) {
		return nil
	}
	if x, ok := m[i].(*VideoMaterial); ok {
		return x
	}
	return nil
}

func (m Materials) GetPDFMaterialElement(i int) *PDFMaterial {
	if i >= len(m) {
		return nil
	}
	if x, ok := m[i].(*PDFMaterial); ok {
		return x
	}
	return nil
}

// Types that are assignable to Materials:
//
//	*VideoMaterial
//	*PDFMaterial
type material interface {
	isMaterialType()
	isValid() error
}

var (
	_ material = new(VideoMaterial)
	_ material = new(PDFMaterial)
)

type VideoMaterial struct {
	ID      string
	Name    string
	VideoID string `json:"media_id"`
}

func (v *VideoMaterial) isMaterialType() {}

func (v *VideoMaterial) isValid() error {
	if len(v.ID) == 0 {
		return fmt.Errorf("ID cannot be empty")
	}

	if len(v.VideoID) == 0 {
		return fmt.Errorf("video ID cannot be empty")
	}

	return nil
}

type PDFMaterial struct {
	ID                string
	Name              string
	URL               string
	ConvertedImageURL *ConvertedImage
}

type ConvertedImage struct {
	Width    int32
	Height   int32
	ImageURL string
}

func (p *PDFMaterial) isMaterialType() {}

func (p *PDFMaterial) isValid() error {
	if len(p.ID) == 0 {
		return fmt.Errorf("ID cannot be empty")
	}

	if len(p.URL) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}

	return nil
}

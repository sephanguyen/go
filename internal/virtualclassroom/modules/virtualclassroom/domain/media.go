package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type MediaType string

const (
	MediaTypeNone           MediaType = "MEDIA_TYPE_NONE"
	MediaTypeVideo          MediaType = "MEDIA_TYPE_VIDEO"
	MediaTypeImage          MediaType = "MEDIA_TYPE_IMAGE"
	MediaTypePDF            MediaType = "MEDIA_TYPE_PDF"
	MediaTypeAudio          MediaType = "MEDIA_TYPE_AUDIO"
	MediaTypeRecordingVideo MediaType = "MEDIA_TYPE_RECORDING_VIDEO"
)

type Media struct {
	ID              string
	Name            string
	Resource        string
	Type            MediaType
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Comments        []Comment
	ConvertedImages []ConvertedImage
	FileSizeBytes   int64
	Duration        time.Duration
}

func (m *Media) PreInsert() {
	m.ID = idutil.ULIDNow()
}

func (m *Media) IsValid() error {
	if len(m.ID) == 0 {
		return fmt.Errorf("Media.ID could not be empty")
	}
	if len(m.Resource) == 0 {
		return fmt.Errorf("Media.Resource could not be empty")
	}
	if len(m.Type) == 0 {
		return fmt.Errorf("Media.Type could not be empty")
	}

	return nil
}

type Medias []*Media

type Comment struct {
	Comment  string
	Duration int64
}

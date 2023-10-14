package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type MediaType string

const (
	MediaTypeNone  MediaType = "MEDIA_TYPE_NONE"
	MediaTypeVideo MediaType = "MEDIA_TYPE_VIDEO"
	MediaTypeImage MediaType = "MEDIA_TYPE_IMAGE"
	MediaTypePDF   MediaType = "MEDIA_TYPE_PDF"
	MediaTypeAudio MediaType = "MEDIA_TYPE_AUDIO"
	MediaTypeZip   MediaType = "MEDIA_TYPE_ZIP"
)

type ConvertedImage struct {
	Width    int32
	Height   int32
	ImageURL string
}

type Media struct {
	MediaID         pgtype.Text
	Name            pgtype.Text
	Resource        pgtype.Text
	Type            pgtype.Text
	Comments        pgtype.JSONB
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
	ConvertedImages pgtype.JSONB // marshal and unmarshal to ConvertedImage struct.
	FileSize        pgtype.Int8
}

func (m *Media) FieldMap() ([]string, []interface{}) {
	fields := []string{
		"media_id",
		"name",
		"resource",
		"converted_images",
		"type", "comments",
		"created_at",
		"updated_at",
		"deleted_at",
		"file_size_bytes",
	}
	return fields, []interface{}{
		&m.MediaID,
		&m.Name,
		&m.Resource,
		&m.ConvertedImages,
		&m.Type, &m.Comments,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.DeletedAt,
		&m.FileSize,
	}
}

func (m *Media) TableName() string {
	return "media"
}

func (m *Media) PreInsert() error {
	if m.MediaID.Status != pgtype.Present {
		if err := m.MediaID.Set(idutil.ULIDNow()); err != nil {
			return err
		}
	}
	now := time.Now()
	err := multierr.Combine(
		m.CreatedAt.Set(now),
		m.UpdatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	return nil
}

type Medias []*Media

func (m *Medias) Add() database.Entity {
	e := &Media{}
	*m = append(*m, e)

	return e
}

func (m Medias) GetUniqueIDs() pgtype.TextArray {
	var ids []string
	for _, i := range m {
		if i.MediaID.Status == pgtype.Present {
			ids = append(ids, i.MediaID.String)
		}
	}
	ids = golibs.GetUniqueElementStringArray(ids)
	res := database.TextArray(ids)

	return res
}

func (m Medias) PreInsert() error {
	for i := range m {
		if err := m[i].PreInsert(); err != nil {
			return err
		}
	}

	return nil
}

func (m Medias) GetUncreatedMedias() Medias {
	var newMedias Medias
	for _, media := range m {
		if media.MediaID.Status != pgtype.Present {
			newMedias = append(newMedias, media)
		}
	}

	return newMedias
}

type Comment struct {
	Comment  string `json:"comment"`
	Duration int64  `json:"duration"`
}

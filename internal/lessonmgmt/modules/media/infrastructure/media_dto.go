package infrastructure

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	mediaDomain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

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
	FileSizeBytes   pgtype.Int8
	DurationSeconds pgtype.Int4
}

func (m *Media) FieldMap() ([]string, []interface{}) {
	return []string{
			"media_id",
			"name",
			"resource",
			"converted_images",
			"type",
			"comments",
			"file_size_bytes",
			"duration_seconds",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&m.MediaID,
			&m.Name,
			&m.Resource,
			&m.ConvertedImages,
			&m.Type,
			&m.Comments,
			&m.FileSizeBytes,
			&m.DurationSeconds,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.DeletedAt,
		}
}

func (m *Media) TableName() string {
	return "media"
}

func (m *Media) ToMediaEntity() *mediaDomain.Media {
	var comments []mediaDomain.Comment
	err := m.Comments.AssignTo(&comments)
	if err != nil {
		return nil
	}
	var convertedImages []mediaDomain.ConvertedImage
	err = m.ConvertedImages.AssignTo(&convertedImages)
	if err != nil {
		return nil
	}

	return &mediaDomain.Media{
		ID:              m.MediaID.String,
		Name:            m.Name.String,
		Resource:        m.Resource.String,
		Type:            mediaDomain.MediaType(m.Type.String),
		CreatedAt:       m.CreatedAt.Time,
		UpdatedAt:       m.UpdatedAt.Time,
		Comments:        comments,
		ConvertedImages: convertedImages,
		FileSizeBytes:   m.FileSizeBytes.Int,
		Duration:        time.Duration(int64(m.DurationSeconds.Int) * time.Second.Nanoseconds()),
	}
}

func NewMediaFromEntity(media *mediaDomain.Media) (*Media, error) {
	dto := &Media{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.MediaID.Set(media.ID),
		dto.Name.Set(media.Name),
		dto.Type.Set(media.Type),
		dto.Resource.Set(media.Resource),
		dto.FileSizeBytes.Set(media.FileSizeBytes),
		dto.DurationSeconds.Set(media.Duration.Seconds()),
		dto.CreatedAt.Set(media.CreatedAt),
		dto.UpdatedAt.Set(media.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not map from media entity to media dto: %w", err)
	}

	return dto, nil
}

type Medias []*Media

func (m *Medias) Add() database.Entity {
	e := &Media{}
	*m = append(*m, e)

	return e
}

func (m Medias) ToMediasEntity() mediaDomain.Medias {
	res := make(mediaDomain.Medias, 0, len(m))
	for _, media := range m {
		var comments []mediaDomain.Comment
		err := media.Comments.AssignTo(&comments)
		if err != nil {
			return nil
		}
		var convertedImages []mediaDomain.ConvertedImage
		err = media.ConvertedImages.AssignTo(&convertedImages)
		if err != nil {
			return nil
		}

		res = append(res, &mediaDomain.Media{
			ID:              media.MediaID.String,
			Name:            media.Name.String,
			Resource:        media.Resource.String,
			Type:            mediaDomain.MediaType(media.Type.String),
			CreatedAt:       media.CreatedAt.Time,
			UpdatedAt:       media.UpdatedAt.Time,
			Comments:        comments,
			ConvertedImages: convertedImages,
			FileSizeBytes:   media.FileSizeBytes.Int,
			Duration:        time.Duration(int64(media.DurationSeconds.Int) * time.Second.Nanoseconds()),
		})
	}

	return res
}

func (m *Media) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		m.CreatedAt.Set(now),
		m.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

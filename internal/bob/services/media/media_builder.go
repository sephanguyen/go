package media

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/media/repo"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Builder struct {
	DB   database.Ext
	repo repo.MediaRepo
}

func NewMediaBuilder(db database.Ext, repo repo.MediaRepo) *Builder {
	return &Builder{
		DB:   db,
		repo: repo,
	}
}

func (s *Builder) Upsert(ctx context.Context, medias entities.Medias) (entities.Medias, error) {
	if len(medias) == 0 {
		return nil, nil
	}

	err := s.repo.UpsertMediaBatch(ctx, s.DB, medias)
	if err != nil {
		return nil, err
	}

	return medias, nil
}

func (s *Builder) RetrieveByIDs(ctx context.Context, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
	if len(mediaIDs.Elements) == 0 {
		return nil, nil
	}

	res, err := s.repo.RetrieveByIDs(ctx, s.DB, mediaIDs)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Builder) CheckMediaIDs(ctx context.Context, mediaIDs pgtype.TextArray) error {
	existingMedia, err := s.RetrieveByIDs(ctx, mediaIDs)
	if err != nil {
		return err
	}
	if len(existingMedia) != len(mediaIDs.Elements) {
		return fmt.Errorf("expect %d media IDs, found %d", len(mediaIDs.Elements), len(existingMedia))
	}

	return nil
}

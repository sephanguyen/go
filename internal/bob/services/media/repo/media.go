package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type MediaRepo interface {
	RetrieveByIDs(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error)
	UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media entities.Medias) error
}

var _ MediaRepo = new(MediaRepoMock)

type MediaRepoMock struct {
	RetrieveByIDsMock    func(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error)
	UpsertMediaBatchMock func(ctx context.Context, db database.QueryExecer, media entities.Medias) error
}

func (m MediaRepoMock) RetrieveByIDs(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error) {
	return m.RetrieveByIDsMock(ctx, db, mediaIDs)
}

func (m MediaRepoMock) UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media entities.Medias) error {
	return m.UpsertMediaBatchMock(ctx, db, media)
}

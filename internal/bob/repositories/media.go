package repositories

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type MediaRepo struct{}

func (m *MediaRepo) QueueMedia(b *pgx.Batch, t *entities_bob.Media) {
	fieldNames := database.GetFieldNames(t)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT media_pk DO UPDATE SET name = $2, resource = $3, type = $5, comments = $6::jsonb, updated_at = $8`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)

	b.Queue(query, database.GetScanFields(t, fieldNames)...)
}

func (m *MediaRepo) UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media entities_bob.Medias) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.UpsertMediaBatch")
	defer span.End()

	if err := media.PreInsert(); err != nil {
		return fmt.Errorf("could not pre-insert for new medidas %v", err)
	}

	b := &pgx.Batch{}
	for _, t := range media {
		m.QueueMedia(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(media); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "UpsertMediaBatch batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("media not inserted")
		}
	}

	return nil
}

func (m *MediaRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities_bob.Media, error) {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.RetrieveByIDs")
	defer span.End()

	e := &entities_bob.Media{}
	fieldNames := database.GetFieldNames(e)

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE media_id = ANY($1)`, strings.Join(fieldNames, ","), e.TableName())

	result := entities_bob.Medias{}
	err := database.Select(ctx, db, query, mediaIDs).ScanAll(&result)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return result, nil
}

func (m *MediaRepo) UpdateConvertedImages(ctx context.Context, db database.QueryExecer, media []*entities_bob.Media) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.UpdateConvertedImages")
	defer span.End()

	b := &pgx.Batch{}

	for _, e := range media {
		query := "UPDATE media SET converted_images = $1, updated_at = NOW() WHERE resource = $2"
		b.Queue(query, &e.ConvertedImages, &e.Resource)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(media); i++ {
		if _, err := batchResults.Exec(); err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

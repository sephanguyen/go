package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/media/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type MediaRepo struct {
}

func (m *MediaRepo) UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media domain.Medias) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.UpsertMediaBatch")
	defer span.End()

	if err := media.PreInsert(); err != nil {
		return fmt.Errorf("could not pre-insert for new medidas %v", err)
	}

	b := &pgx.Batch{}
	for _, t := range media {
		m.queueMedia(b, t)
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

func (m *MediaRepo) queueMedia(b *pgx.Batch, t *domain.Media) {
	fieldNames := database.GetFieldNames(t)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES (%s)
		ON CONFLICT ON CONSTRAINT media_pk 
		DO UPDATE SET 
			name = $2, 
			resource = $3, 
			type = $5, 
			comments = $6::jsonb, 
			updated_at = $8,
			deleted_at = NULL
	`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)

	b.Queue(query, database.GetScanFields(t, fieldNames)...)
}

package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type MediaRepo struct{}

func (m *MediaRepo) InsertMedia(ctx context.Context, db database.QueryExecer, media *domain.Media) (*domain.Media, error) {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.InsertMedia")
	defer span.End()

	dto, err := NewMediaFromEntity(media)
	if err != nil {
		return nil, err
	}
	if err = dto.PreInsert(); err != nil {
		return nil, fmt.Errorf("got error when PreInsert recorded video dto: %w", err)
	}

	fieldNames, args := dto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO media (%s) VALUES (%s)",
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err = db.Exec(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("media id %s: %v", media.ID, err)
	}

	media.CreatedAt = dto.CreatedAt.Time
	media.UpdatedAt = dto.UpdatedAt.Time
	return media, nil
}

func (m *MediaRepo) ListByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) (domain.Medias, error) {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.ListByIDs")
	defer span.End()

	e := &Media{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE media_id = ANY($1) AND deleted_at IS NULL`, strings.Join(fieldNames, ","), e.TableName())
	result := Medias{}
	err := database.Select(ctx, db, query, mediaIDs).ScanAll(&result)
	if err != nil {
		return nil, err
	}
	return result.ToMediasEntity(), nil
}

func (m *MediaRepo) DeleteByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.DeleteByIDs")
	defer span.End()

	query := "UPDATE media SET deleted_at = now(), updated_at = now() WHERE media_id = ANY($1) AND deleted_at IS NULL"
	command, err := db.Exec(ctx, query, &mediaIDs)
	if command.RowsAffected() == 0 {
		return fmt.Errorf("not found any media to delete")
	}

	return err
}

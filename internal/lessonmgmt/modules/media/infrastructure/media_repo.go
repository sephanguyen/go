package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type MediaRepo struct{}

func (m *MediaRepo) RetrieveMediasByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) (domain.Medias, error) {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.RetrieveMediasByIDs")
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

func (m *MediaRepo) CreateMedia(ctx context.Context, db database.QueryExecer, media *domain.Media) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.CreateMedia")
	defer span.End()

	dto, err := NewMediaFromEntity(media)
	if err != nil {
		return err
	}
	if err = dto.PreInsert(); err != nil {
		return fmt.Errorf("got error when PreInsert media dto: %w", err)
	}

	fieldNames, args := dto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO media (%s) VALUES (%s)",
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err = db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("db.Exec, media id %s: %w", media.ID, err)
	}

	return nil
}

func (m *MediaRepo) DeleteMedias(ctx context.Context, db database.QueryExecer, mediaIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "MediaRepo.DeleteMedias")
	defer span.End()

	query := `UPDATE media 
		SET deleted_at = now(), updated_at = now() 
		WHERE media_id = ANY($1) 
		AND deleted_at IS NULL`
	command, err := db.Exec(ctx, query, &mediaIDs)
	if err != nil {
		return fmt.Errorf("db.Exec, media ids %v: %w", mediaIDs, err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("not found any media to delete")
	}

	return err
}

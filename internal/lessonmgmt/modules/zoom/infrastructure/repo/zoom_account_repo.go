package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type ZoomAccountRepo struct{}

func (l *ZoomAccountRepo) GetZoomAccountByID(ctx context.Context, db database.QueryExecer, id string) (*domain.ZoomAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "ZoomAccountRepo.GetZoomAccountById")
	defer span.End()

	zoomAccount := &ZoomAccount{}
	fields, values := zoomAccount.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM zoom_account
		WHERE zoom_id = $1 AND deleted_at IS NULL `,
		strings.Join(fields, ","),
	)
	if err := db.QueryRow(ctx, query, &id).Scan(values...); err != nil {
		return nil, err
	}

	return zoomAccount.ToZoomAccountEntity(), nil
}

func (l *ZoomAccountRepo) Upsert(ctx context.Context, db database.Ext, zoomAccounts domain.ZoomAccounts) error {
	ctx, span := interceptors.StartSpan(ctx, "ZoomAccountRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}

	for _, zoomAccount := range zoomAccounts {
		detailDTO := &ZoomAccount{}
		database.AllNullEntity(detailDTO)
		if err := multierr.Combine(
			detailDTO.ID.Set(zoomAccount.ID),
			detailDTO.Email.Set(zoomAccount.Email),
			detailDTO.UserName.Set(zoomAccount.UserName),
			detailDTO.CreatedAt.Set(zoomAccount.CreatedAt),
			detailDTO.UpdatedAt.Set(zoomAccount.UpdatedAt),
			detailDTO.DeletedAt.Set(zoomAccount.DeletedAt),
		); err != nil {
			return fmt.Errorf("could not mapping from zoom account entity to zoom account dto: %w", err)
		}
		l.UpsertQueue(b, detailDTO)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *ZoomAccountRepo) UpsertQueue(b *pgx.Batch, e *ZoomAccount) {
	fields, values := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT pk__zoom_account DO "+
		"UPDATE SET email = $2, user_name = $3, updated_at = now(), deleted_at = $6", e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (l *ZoomAccountRepo) GetAllZoomAccount(ctx context.Context, db database.QueryExecer) ([]*ZoomAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "ZoomAccountRepo.GetAllZoomAccount")
	defer span.End()

	zoomAccount := &ZoomAccount{}
	fields, _ := zoomAccount.FieldMap()

	query := fmt.Sprintf(`SELECT %s 
		FROM %s WHERE deleted_at IS NULL`, strings.Join(fields, ","),
		zoomAccount.TableName())

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all date info: db.Query")
	}
	defer rows.Close()

	allZoomAccount := []*ZoomAccount{}
	for rows.Next() {
		item := &ZoomAccount{}
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		allZoomAccount = append(allZoomAccount, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to get all zoom Account rows.Err")
	}

	return allZoomAccount, nil
}

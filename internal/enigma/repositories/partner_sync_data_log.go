package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"go.uber.org/multierr"
)

type PartnerSyncDataLogRepo struct{}

func (p *PartnerSyncDataLogRepo) Create(ctx context.Context, db database.QueryExecer, log *entities.PartnerSyncDataLog) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogRepo.Create")
	defer span.End()

	now := timeutil.Now()
	if err := multierr.Combine(
		log.CreatedAt.Set(now),
		log.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	if _, err := database.InsertIgnoreConflict(ctx, log, db.Exec); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (p *PartnerSyncDataLogRepo) GetBySignature(ctx context.Context, db database.QueryExecer, signature string) (*entities.PartnerSyncDataLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogRepo.GetBySignature")
	defer span.End()

	log := &entities.PartnerSyncDataLog{}
	fields, values := log.FieldMap()
	query := fmt.Sprintf(`
			SELECT %s FROM %s
			WHERE signature = $1`,
		strings.Join(fields, ","), log.TableName(),
	)

	err := db.QueryRow(ctx, query, &signature).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return log, nil
}

func (p *PartnerSyncDataLogRepo) UpdateTime(ctx context.Context, db database.QueryExecer, logID string) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogRepo.UpdateTime")
	defer span.End()
	log := &entities.PartnerSyncDataLog{}
	updateLogStatusStmt := fmt.Sprintf(`UPDATE %s SET updated_at = now() WHERE partner_sync_data_log_id = $1`, log.TableName())
	cmdTag, err := db.Exec(ctx, updateLogStatusStmt, logID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

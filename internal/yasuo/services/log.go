package services

import (
	"context"
	"fmt"

	entities_enigma "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

type LogsService struct {
	DB database.Ext

	PartnerSyncDataLogSplitRepo interface {
		UpdateLogStatus(ctx context.Context, db database.QueryExecer, log *entities_enigma.PartnerSyncDataLogSplit) error
	}
}

func (l *LogsService) UpdateLogStatus(ctx context.Context, id, status string) error {
	partnerLogSplit := &entities_enigma.PartnerSyncDataLogSplit{}
	database.AllNullEntity(partnerLogSplit)
	if err := multierr.Combine(
		partnerLogSplit.PartnerSyncDataLogSplitID.Set(id),
		partnerLogSplit.Status.Set(status),
	); err != nil {
		return fmt.Errorf("LogsService.UpdateLogStatus.Combine: %w", err)
	}

	if err := l.PartnerSyncDataLogSplitRepo.UpdateLogStatus(ctx, l.DB, partnerLogSplit); err != nil {
		return fmt.Errorf("LogsService.UpdateLogStatus: %w", err)
	}

	return nil
}

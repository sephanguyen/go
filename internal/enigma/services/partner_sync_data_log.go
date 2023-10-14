package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

type PartnerSyncDataLogService struct {
	DB database.Ext

	PartnerSyncDataLogRepo interface {
		GetBySignature(ctx context.Context, db database.QueryExecer, signature string) (*entities.PartnerSyncDataLog, error)
	}

	PartnerSyncDataLogSplitRepo interface {
		UpdateLogStatus(ctx context.Context, db database.QueryExecer, log *entities.PartnerSyncDataLogSplit) error
	}
}

var (
	ErrPartnerSyncDataLogStatusUnknown = fmt.Errorf("partner sync data log status unknown")
)

func (p *PartnerSyncDataLogService) UpdateLogStatus(ctx context.Context, id, status string) error {
	if len(id) == 0 {
		return nil
	}

	if _, ok := entities.PartnerSyncDataLogStatusValue[status]; !ok {
		return fmt.Errorf("PartnerSyncDataLogService.UpdateLogStatus: %s", ErrPartnerSyncDataLogStatusUnknown)
	}

	partnerLogSplit := &entities.PartnerSyncDataLogSplit{}
	database.AllNullEntity(partnerLogSplit)
	if err := multierr.Combine(
		partnerLogSplit.PartnerSyncDataLogSplitID.Set(id),
		partnerLogSplit.Status.Set(status),
	); err != nil {
		return fmt.Errorf("PartnerSyncDataLogService.UpdateLogStatus.Combine: %w", err)
	}

	if err := p.PartnerSyncDataLogSplitRepo.UpdateLogStatus(ctx, p.DB, partnerLogSplit); err != nil {
		return fmt.Errorf("PartnerSyncDataLogService.UpdateLogStatus: %w", err)
	}

	return nil
}

func (p *PartnerSyncDataLogService) GetLogBySignature(ctx context.Context, signature string) (*entities.PartnerSyncDataLog, error) {
	if len(signature) == 0 {
		return nil, fmt.Errorf("Signature is empty")
	}
	return p.PartnerSyncDataLogRepo.GetBySignature(ctx, p.DB, signature)
}

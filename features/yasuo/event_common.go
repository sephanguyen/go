package yasuo

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entites "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"

	"go.uber.org/multierr"
)

func (s *suite) createPartnerSyncDataLog(ctx context.Context, signature string, hours time.Duration) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	partnerSyncDataLog := &enigma_entites.PartnerSyncDataLog{}
	now := time.Now()
	newPartnerSyncDataLogId := idutil.ULIDNow()
	stepState.PartnerSyncDataLogId = newPartnerSyncDataLogId

	err := multierr.Combine(
		partnerSyncDataLog.PartnerSyncDataLogID.Set(newPartnerSyncDataLogId),
		partnerSyncDataLog.Signature.Set(signature),
		partnerSyncDataLog.Payload.Set([]byte("{}")),
		partnerSyncDataLog.CreatedAt.Set(now.Add(-hours*time.Hour)),
		partnerSyncDataLog.UpdatedAt.Set(now.Add(-hours*time.Hour)),
	)

	if _, err = database.InsertIgnoreConflict(ctx, partnerSyncDataLog, db.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert partner sync data log err: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLogSyncDataSplit(ctx context.Context, kind string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	partnerSyncDataLogSplit := &enigma_entites.PartnerSyncDataLogSplit{}

	database.AllNullEntity(partnerSyncDataLogSplit)
	now := time.Now()

	newPartnerSyncDataLogSplitId := idutil.ULIDNow()
	stepState.PartnerSyncDataLogSplitId = newPartnerSyncDataLogSplitId
	err := multierr.Combine(
		partnerSyncDataLogSplit.PartnerSyncDataLogSplitID.Set(newPartnerSyncDataLogSplitId),
		partnerSyncDataLogSplit.PartnerSyncDataLogID.Set(stepState.PartnerSyncDataLogId),
		partnerSyncDataLogSplit.Payload.Set([]byte("{}")),
		partnerSyncDataLogSplit.Kind.Set(kind),
		partnerSyncDataLogSplit.Status.Set(string(enigma_entites.StatusPending)),
		partnerSyncDataLogSplit.RetryTimes.Set(0),
		partnerSyncDataLogSplit.CreatedAt.Set(now),
		partnerSyncDataLogSplit.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if _, err = database.InsertIgnoreConflict(ctx, partnerSyncDataLogSplit, db.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert partner sync data log split id err: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) storeLogDataSplitWithCorrectStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var statusFromDB string
	var count int
	query := `SELECT status FROM partner_sync_data_log_split
		WHERE partner_sync_data_log_id = $1 AND partner_sync_data_log_split_id = $2 LIMIT 1`
	err := try.Do(func(attempt int) (bool, error) {
		err := s.DB.QueryRow(ctx, query, stepState.PartnerSyncDataLogId, stepState.PartnerSyncDataLogSplitId).Scan(&statusFromDB)
		if err == nil && count > 0 {
			return false, nil
		}
		if statusFromDB == "PROCESSING" && count > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(2 * time.Second)
			return true, fmt.Errorf("error status log: %w", err)
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if statusFromDB != status && statusFromDB != "PROCESSING" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected status data log split expect %s but status in database is %s", status, statusFromDB)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) storeLastTimeRecievedMessageCorrect(ctx context.Context, configKey string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	configRepo := repositories.ConfigRepo{}
	configs, err := configRepo.Retrieve(ctx, s.DBTrace, database.Text("COUNTRY_MASTER"), database.Text("yasuo"), database.TextArray([]string{configKey}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("query configRepo err: %w", err)
	}
	if len(configs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't save last time received message")
	}
	return StepStateToContext(ctx, stepState), nil
}

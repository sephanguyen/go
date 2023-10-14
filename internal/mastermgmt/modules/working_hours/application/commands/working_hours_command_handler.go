package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ImportWorkingHoursCsvFields struct {
	Day         string
	OpeningTime string
	ClosingTime string
}

type WorkingHoursCommandHandler struct {
	DB               database.Ext
	WorkingHoursRepo infrastructure.WorkingHoursRepo
}

func (wch *WorkingHoursCommandHandler) ImportWorkingHoursTx(ctx context.Context, payload ImportWorkingHoursPayload, locationIDs []string) error {
	err := database.ExecInTx(ctx, wch.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := wch.WorkingHoursRepo.Upsert(ctx, tx, payload.WorkingHours, locationIDs)
		return err
	})
	if err != nil {
		return fmt.Errorf("ImportWorkingHoursTx: %w", err)
	}
	return nil
}

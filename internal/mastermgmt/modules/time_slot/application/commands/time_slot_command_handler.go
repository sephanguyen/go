package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ImportTimeSlotCsvFields struct {
	TimeSlotInternalID string
	StartTime          string
	EndTime            string
}

type TimeSlotCommandHandler struct {
	DB           database.Ext
	TimeSlotRepo infrastructure.TimeSlotRepo
}

func (tsc *TimeSlotCommandHandler) ImportTimeSlotTx(ctx context.Context, payload ImportTimeSlotPayload, locationIDs []string) error {
	err := database.ExecInTx(ctx, tsc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := tsc.TimeSlotRepo.Upsert(ctx, tx, payload.TimeSlots, locationIDs)
		return err
	})
	if err != nil {
		return fmt.Errorf("ImportTimeSlotTx: %w", err)
	}
	return nil
}

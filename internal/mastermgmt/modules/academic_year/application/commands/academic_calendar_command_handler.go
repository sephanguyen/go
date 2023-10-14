package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type ImportAcademicCalendarCsvFields struct {
	Order                 int
	AcademicWeekName      string
	AcademicWeekStartDate time.Time
	AcademicWeekEndDate   time.Time
	Period                string
	AcademicClosedDays    []time.Time
}

type AcademicCalendarCommandHandler struct {
	DB                    database.Ext
	AcademicWeekRepo      infrastructure.AcademicWeekRepo
	AcademicClosedDayRepo infrastructure.AcademicClosedDayRepo
}

func (a *AcademicCalendarCommandHandler) ImportAcademicCalendarTx(ctx context.Context, payload ImportAcademicCalendarPayload) error {
	err := database.ExecInTx(ctx, a.DB, func(ctx context.Context, tx pgx.Tx) error {
		err1 := a.AcademicWeekRepo.Insert(ctx, tx, payload.AcademicWeeks)
		err2 := a.AcademicClosedDayRepo.Insert(ctx, tx, payload.AcademicClosedDays)
		combineErr := multierr.Combine(err1, err2)
		return combineErr
	})
	if err != nil {
		return fmt.Errorf("ImportAcademicCalendarTx: %w", err)
	}
	return nil
}

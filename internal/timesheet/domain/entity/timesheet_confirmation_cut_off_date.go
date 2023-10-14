package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type TimesheetConfirmationCutOffDate struct {
	ID         pgtype.Text
	CutOffDate pgtype.Int4
	StartDate  pgtype.Timestamptz
	EndDate    pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (t *TimesheetConfirmationCutOffDate) TableName() string {
	return "timesheet_confirmation_cut_off_date"
}

func NewCutOffDate() *TimesheetConfirmationCutOffDate {
	return &TimesheetConfirmationCutOffDate{
		ID:        database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}
}

func (t *TimesheetConfirmationCutOffDate) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"cut_off_date",
			"start_date",
			"end_date",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.CutOffDate,
			&t.StartDate,
			&t.EndDate,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

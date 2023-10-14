package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Timesheet struct {
	TimesheetID     pgtype.Text
	StaffID         pgtype.Text
	LocationID      pgtype.Text
	TimesheetStatus pgtype.Text
	TimesheetDate   pgtype.Timestamptz
	Remark          pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

func NewTimesheet() *Timesheet {
	return &Timesheet{
		TimesheetID:     database.Text(idutil.ULIDNow()),
		Remark:          pgtype.Text{Status: pgtype.Null},
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
		DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
	}
}

func (t *Timesheet) FieldMap() ([]string, []interface{}) {
	return []string{
			"timesheet_id", "staff_id", "location_id", "timesheet_status", "timesheet_date", "remark", "created_at", "updated_at", "deleted_at",
		}, []interface{}{
			&t.TimesheetID, &t.StaffID, &t.LocationID, &t.TimesheetStatus, &t.TimesheetDate, &t.Remark, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		}
}

func (t *Timesheet) TableName() string {
	return "timesheet"
}

func (t *Timesheet) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		t.CreatedAt.Set(now),
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}

func (t *Timesheet) PreUpdate() error {
	now := time.Now()
	return multierr.Combine(
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}

type Timesheets []*Timesheet

// Add append new Timesheet
func (t *Timesheets) Add() database.Entity {
	e := &Timesheet{}
	*t = append(*t, e)

	return e
}

func (t *Timesheet) PrimaryKey() string {
	return "timesheet_id"
}

func (t *Timesheet) UpdateOnConflictQuery() string {
	return `
	timesheet_status = EXCLUDED.timesheet_status,
	remark = EXCLUDED.remark,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

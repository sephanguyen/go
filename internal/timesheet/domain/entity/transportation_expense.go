package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type TransportationExpense struct {
	TransportationExpenseID pgtype.Text
	TimesheetID             pgtype.Text
	TransportationType      pgtype.Text
	TransportationFrom      pgtype.Text
	TransportationTo        pgtype.Text
	CostAmount              pgtype.Int4
	RoundTrip               pgtype.Bool
	Remarks                 pgtype.Text
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (t *TransportationExpense) FieldMap() ([]string, []interface{}) {
	return []string{
			"transportation_expense_id",
			"timesheet_id",
			"transportation_type",
			"transportation_from",
			"transportation_to",
			"cost_amount",
			"round_trip",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.TransportationExpenseID,
			&t.TimesheetID,
			&t.TransportationType,
			&t.TransportationFrom,
			&t.TransportationTo,
			&t.CostAmount,
			&t.RoundTrip,
			&t.Remarks,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

func (t *TransportationExpense) TableName() string {
	return "transportation_expense"
}

func (*TransportationExpense) PrimaryField() string {
	return "transportation_expense_id"
}

func (*TransportationExpense) UpsertConflictField() string {
	return "transportation_expense_id"
}

func (t *TransportationExpense) UpdateOnConflictQuery() string {
	return `
	transportation_type = EXCLUDED.transportation_type,
	transportation_from = EXCLUDED.transportation_from,
	transportation_to = EXCLUDED.transportation_to,
	cost_amount = EXCLUDED.cost_amount,
	round_trip = EXCLUDED.round_trip,
	remarks = EXCLUDED.remarks,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

func NewTransportExpenses() *TransportationExpense {
	return &TransportationExpense{
		TransportationExpenseID: database.Text(idutil.ULIDNow()),
		TransportationFrom:      pgtype.Text{Status: pgtype.Null},
		TransportationTo:        pgtype.Text{Status: pgtype.Null},
		Remarks:                 pgtype.Text{Status: pgtype.Null},
		DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
	}
}

type ListTransportationExpenses []*TransportationExpense

// Add append new Timesheet
func (lt *ListTransportationExpenses) Add() database.Entity {
	t := &TransportationExpense{}
	*lt = append(*lt, t)

	return t
}

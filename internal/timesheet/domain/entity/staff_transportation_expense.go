package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type StaffTransportationExpense struct {
	ID                 pgtype.Text
	StaffID            pgtype.Text
	LocationID         pgtype.Text
	TransportationType pgtype.Text
	TransportationFrom pgtype.Text
	TransportationTo   pgtype.Text
	CostAmount         pgtype.Int4
	RoundTrip          pgtype.Bool
	Remarks            pgtype.Text
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
}

func (t *StaffTransportationExpense) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"staff_id",
			"location_id",
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
			&t.ID,
			&t.StaffID,
			&t.LocationID,
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

func (t *StaffTransportationExpense) TableName() string {
	return "staff_transportation_expense"
}

func (*StaffTransportationExpense) PrimaryField() string {
	return "id"
}

func (*StaffTransportationExpense) UpsertConflictField() string {
	return "id"
}

func (t *StaffTransportationExpense) UpdateOnConflictQuery() string {
	return `
	staff_id = EXCLUDED.staff_id,
	location_id = EXCLUDED.location_id,
	transportation_type = EXCLUDED.transportation_type,
	transportation_from = EXCLUDED.transportation_from,
	transportation_to = EXCLUDED.transportation_to,
	cost_amount = EXCLUDED.cost_amount,
	round_trip = EXCLUDED.round_trip,
	remarks = EXCLUDED.remarks,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

func NewStaffTransportationExpense() *StaffTransportationExpense {
	return &StaffTransportationExpense{
		ID:                 database.Text(idutil.ULIDNow()),
		TransportationFrom: pgtype.Text{Status: pgtype.Null},
		TransportationTo:   pgtype.Text{Status: pgtype.Null},
		Remarks:            pgtype.Text{Status: pgtype.Null},
		DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
	}
}

type ListStaffTransportationExpense []*StaffTransportationExpense

// Add append new Timesheet
func (lt *ListStaffTransportationExpense) Add() database.Entity {
	t := &StaffTransportationExpense{}
	*lt = append(*lt, t)

	return t
}

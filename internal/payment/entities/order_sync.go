package entities

import "github.com/jackc/pgtype"

type (
	OrderSync struct {
		ID                  pgtype.Text
		StudentID           pgtype.Text
		StudentFullName     pgtype.Text
		LocationID          pgtype.Text
		OrderSequenceNumber pgtype.Int4
		OrderComment        pgtype.Text
		OrderStatus         pgtype.Text
		OrderType           pgtype.Text
		UpdatedAt           pgtype.Timestamptz
		CreatedAt           pgtype.Timestamptz
		OrderItems          []*OrderItemSync
		ResourcePath        pgtype.Text
		IsReviewed          pgtype.Bool
	}

	OrderItemSync struct {
		DiscountID   pgtype.Text
		StartDate    pgtype.Timestamptz
		CreatedAt    pgtype.Timestamptz
		Product      *ProductSync
		ResourcePath pgtype.Text
	}

	ProductSync struct {
		ID                   pgtype.Text
		Name                 pgtype.Text
		ProductType          pgtype.Text
		TaxID                pgtype.Text
		AvailableFrom        pgtype.Timestamptz
		AvailableUntil       pgtype.Timestamptz
		CustomBillingPeriod  pgtype.Timestamptz
		BillingScheduleID    pgtype.Text
		DisableProRatingFlag pgtype.Bool
		Remarks              pgtype.Text
		IsArchived           pgtype.Bool
		UpdatedAt            pgtype.Timestamptz
		CreatedAt            pgtype.Timestamptz
		ResourcePath         pgtype.Text
	}
)

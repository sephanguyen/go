package entities

import "github.com/jackc/pgtype"

type OrderStats struct {
	TotalItems          pgtype.Int8
	TotalOfSubmitted    pgtype.Int8
	TotalOfPending      pgtype.Int8
	TotalOfRejected     pgtype.Int8
	TotalOfVoided       pgtype.Int8
	TotalOfInvoiced     pgtype.Int8
	TotalOfNeedToReview pgtype.Int8
}

func (e *OrderStats) FieldOrderStatsMap() ([]string, []interface{}) {
	return []string{
			"total_items",
			"total_of_submitted",
			"total_of_pending",
			"total_of_rejected",
			"total_of_voided",
			"total_of_invoiced",
			"total_of_order_need_to_review",
		}, []interface{}{
			&e.TotalItems,
			&e.TotalOfSubmitted,
			&e.TotalOfPending,
			&e.TotalOfRejected,
			&e.TotalOfVoided,
			&e.TotalOfInvoiced,
			&e.TotalOfNeedToReview,
		}
}

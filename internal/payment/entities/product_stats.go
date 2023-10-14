package entities

import "github.com/jackc/pgtype"

type ProductStats struct {
	TotalItems      pgtype.Int8
	TotalOfActive   pgtype.Int8
	TotalOfInactive pgtype.Int8
}

func (e *ProductStats) FieldProductStatsMap() ([]string, []interface{}) {
	return []string{
			"total_items",
			"total_of_active",
			"total_of_inactive",
		}, []interface{}{
			&e.TotalItems,
			&e.TotalOfActive,
			&e.TotalOfInactive,
		}
}

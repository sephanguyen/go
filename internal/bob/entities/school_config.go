package entities

import "github.com/jackc/pgtype"

type SchoolConfig struct {
	ID            pgtype.Int4 `sql:"school_id,pk"`
	PlanID        pgtype.Text `sql:"plan_id"`
	Country       pgtype.Text
	PlanExpiredAt pgtype.Timestamptz
	PlanDuration  pgtype.Int2
	Privileges    pgtype.TextArray
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (t *SchoolConfig) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_id", "plan_id", "country", "plan_expired_at", "plan_duration", "privileges", "updated_at", "created_at",
		}, []interface{}{
			&t.ID, &t.PlanID, &t.Country, &t.PlanExpiredAt, &t.PlanDuration, &t.Privileges, &t.UpdatedAt, &t.CreatedAt,
		}
}

func (t *SchoolConfig) TableName() string {
	return "school_configs"
}

package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type StudentSubscriptionAccessPath struct {
	StudentSubscriptionID pgtype.Text
	LocationID            pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
}

// FieldMap Student Subscription Access Path table data fields
func (ss *StudentSubscriptionAccessPath) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_subscription_id",
			"location_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&ss.StudentSubscriptionID,
			&ss.LocationID,
			&ss.CreatedAt,
			&ss.UpdatedAt,
			&ss.DeletedAt,
		}
}

// TableName returns "lesson_student_subscription_access_path"
func (ss *StudentSubscriptionAccessPath) TableName() string {
	return "lesson_student_subscription_access_path"
}

func (ss *StudentSubscriptionAccessPath) PreUpdate() error {
	return ss.UpdatedAt.Set(time.Now())
}

type StudentSubscriptionAccessPaths []*StudentSubscriptionAccessPath

func (ss *StudentSubscriptionAccessPaths) Add() database.Entity {
	e := &StudentSubscriptionAccessPath{}
	*ss = append(*ss, e)

	return e
}

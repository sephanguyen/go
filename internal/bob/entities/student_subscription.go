package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// Student Subscription reflects lesson_student_subscriptions table
type StudentSubscription struct {
	StudentSubscriptionID pgtype.Text
	CourseID              pgtype.Text
	StudentID             pgtype.Text
	SubscriptionID        pgtype.Text
	StartAt               pgtype.Timestamptz
	EndAt                 pgtype.Timestamptz
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
	PurchasedSlotTotal    pgtype.Int4
}

// FieldMap Student Subscription table data fields
func (ss *StudentSubscription) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_subscription_id",
			"course_id",
			"student_id",
			"subscription_id",
			"start_at",
			"end_at",
			"created_at",
			"updated_at",
			"deleted_at",
			"purchased_slot_total",
		}, []interface{}{
			&ss.StudentSubscriptionID,
			&ss.CourseID,
			&ss.StudentID,
			&ss.SubscriptionID,
			&ss.StartAt,
			&ss.EndAt,
			&ss.CreatedAt,
			&ss.UpdatedAt,
			&ss.DeletedAt,
			&ss.PurchasedSlotTotal,
		}
}

// TableName returns "lesson_student_subscriptions"
func (ss *StudentSubscription) TableName() string {
	return "lesson_student_subscriptions"
}

func (ss *StudentSubscription) PreUpdate() error {
	return ss.UpdatedAt.Set(time.Now())
}

type StudentSubscriptions []*StudentSubscription

func (ss *StudentSubscriptions) Add() database.Entity {
	e := &StudentSubscription{}
	*ss = append(*ss, e)

	return e
}

package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudentSubscriptionAccessPath struct {
	StudentSubscriptionID pgtype.Text
	LocationID            pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
}

func NewStudentSubscriptionAccessPathListFromDomainList(studentSubList domain.StudentSubscriptionAccessPaths) (StudentSubscriptionAccessPaths, error) {
	dtoList := make(StudentSubscriptionAccessPaths, 0, len(studentSubList))

	for _, studentSubAccessPath := range studentSubList {
		dto := &StudentSubscriptionAccessPath{}
		database.AllNullEntity(dto)
		if err := multierr.Combine(
			dto.StudentSubscriptionID.Set(studentSubAccessPath.SubscriptionID),
			dto.LocationID.Set(studentSubAccessPath.LocationID),
		); err != nil {
			return nil, fmt.Errorf("could not map student subscription access path: %w", err)
		}

		dtoList = append(dtoList, dto)
	}

	return dtoList, nil
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

func (ss *StudentSubscriptionAccessPath) PreUpsert() error {
	now := time.Now()

	if err := multierr.Combine(
		ss.CreatedAt.Set(now),
		ss.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	return nil
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

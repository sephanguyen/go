package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudentSubscription struct {
	StudentSubscriptionID pgtype.Text
	CourseID              pgtype.Text
	StudentID             pgtype.Text
	SubscriptionID        pgtype.Text
	StartAt               pgtype.Timestamptz
	EndAt                 pgtype.Timestamptz
	CourseSlot            pgtype.Int4
	CourseSlotPerWeek     pgtype.Int4
	StudentFirstName      pgtype.Text
	StudentLastName       pgtype.Text
	PackageType           pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
}

func NewStudentSubscriptionListFromDomainList(studentSubList domain.StudentSubscriptions) (StudentSubscriptions, error) {
	dtoList := make(StudentSubscriptions, 0, len(studentSubList))

	for _, studentSubInfo := range studentSubList {
		dto := &StudentSubscription{}
		database.AllNullEntity(dto)
		if err := multierr.Combine(
			dto.StudentSubscriptionID.Set(studentSubInfo.StudentSubscriptionID),
			dto.CourseID.Set(studentSubInfo.CourseID),
			dto.StudentID.Set(studentSubInfo.StudentID),
			dto.SubscriptionID.Set(studentSubInfo.SubscriptionID),
			dto.StartAt.Set(studentSubInfo.StartAt),
			dto.EndAt.Set(studentSubInfo.EndAt),
			dto.CourseSlot.Set(studentSubInfo.CourseSlot),
			dto.CourseSlotPerWeek.Set(studentSubInfo.CourseSlotPerWeek),
			dto.StudentFirstName.Set(studentSubInfo.StudentFirstName),
			dto.StudentLastName.Set(studentSubInfo.StudentLastName),
			dto.PackageType.Set(studentSubInfo.PackageType),
		); err != nil {
			return nil, fmt.Errorf("could not map student subscription: %w", err)
		}

		dtoList = append(dtoList, dto)
	}

	return dtoList, nil
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
			"course_slot",
			"course_slot_per_week",
			"student_first_name",
			"student_last_name",
			"package_type",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&ss.StudentSubscriptionID,
			&ss.CourseID,
			&ss.StudentID,
			&ss.SubscriptionID,
			&ss.StartAt,
			&ss.EndAt,
			&ss.CourseSlot,
			&ss.CourseSlotPerWeek,
			&ss.StudentFirstName,
			&ss.StudentLastName,
			&ss.PackageType,
			&ss.CreatedAt,
			&ss.UpdatedAt,
			&ss.DeletedAt,
		}
}

// TableName returns "lesson_student_subscriptions"
func (ss *StudentSubscription) TableName() string {
	return "lesson_student_subscriptions"
}

func (ss *StudentSubscription) PreUpsert() error {
	now := time.Now()

	if err := multierr.Combine(
		ss.CreatedAt.Set(now),
		ss.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	return nil
}

func (ss *StudentSubscription) PreUpdate() error {
	return ss.UpdatedAt.Set(time.Now())
}

func (ss *StudentSubscription) ToStudentSubscriptionEntity() *domain.StudentSubscription {
	domain := &domain.StudentSubscription{
		StudentSubscriptionID: ss.StudentSubscriptionID.String,
		SubscriptionID:        ss.StudentSubscriptionID.String,
		StudentID:             ss.StudentID.String,
		CourseID:              ss.CourseID.String,
		StartAt:               ss.StartAt.Time,
		EndAt:                 ss.EndAt.Time,
		CreatedAt:             ss.CreatedAt.Time,
		UpdatedAt:             ss.UpdatedAt.Time,
	}
	return domain
}

type StudentSubscriptions []*StudentSubscription

func (ss StudentSubscriptions) ToListStudentSubscriptionEntities(locationsBySubscriptionID map[string][]string, gradesBySubscriptionID map[string]string) domain.StudentSubscriptions {
	res := make(domain.StudentSubscriptions, 0, len(ss))
	for _, v := range ss {
		locations := locationsBySubscriptionID[v.StudentSubscriptionID.String]
		grade := gradesBySubscriptionID[v.StudentSubscriptionID.String]
		res = append(res, &domain.StudentSubscription{
			SubscriptionID: v.StudentSubscriptionID.String,
			StudentID:      v.StudentID.String,
			CourseID:       v.CourseID.String,
			StartAt:        v.StartAt.Time,
			EndAt:          v.EndAt.Time,
			CreatedAt:      v.CreatedAt.Time,
			UpdatedAt:      v.UpdatedAt.Time,
			LocationIDs:    locations,
			GradeV2:        grade,
		})
	}

	return res
}

func (ss *StudentSubscriptions) Add() database.Entity {
	e := &StudentSubscription{}
	*ss = append(*ss, e)

	return e
}

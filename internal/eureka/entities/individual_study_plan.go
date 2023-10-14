package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type IndividualStudyPlan struct {
	BaseEntity
	ID                 pgtype.Text
	LearningMaterialID pgtype.Text
	StudentID          pgtype.Text
	AvailableFrom      pgtype.Timestamptz
	AvailableTo        pgtype.Timestamptz
	StartDate          pgtype.Timestamptz
	EndDate            pgtype.Timestamptz
	Status             pgtype.Text
	SchoolDate         pgtype.Timestamptz
}

func (e *IndividualStudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"study_plan_id",
		"learning_material_id",
		"student_id",
		"available_from",
		"available_to",
		"start_date",
		"end_date",
		"updated_at",
		"created_at",
		"deleted_at",
		"status",
		"school_date",
	}
	values = []interface{}{
		&e.ID,
		&e.LearningMaterialID,
		&e.StudentID,
		&e.AvailableFrom,
		&e.AvailableTo,
		&e.StartDate,
		&e.EndDate,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.Status,
		&e.SchoolDate,
	}
	return
}

func (e *IndividualStudyPlan) TableName() string {
	return "individual_study_plan"
}

type IndividualStudyPlans []*IndividualStudyPlan

func (e *IndividualStudyPlans) Add() database.Entity {
	u := &IndividualStudyPlan{}
	*e = append(*e, u)

	return u
}

func (e IndividualStudyPlans) Get() []*IndividualStudyPlan {
	return []*IndividualStudyPlan(e)
}

package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type MasterStudyPlan struct {
	BaseEntity
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	Status             pgtype.Text
	StartDate          pgtype.Timestamptz
	EndDate            pgtype.Timestamptz
	AvailableFrom      pgtype.Timestamptz
	AvailableTo        pgtype.Timestamptz
	SchoolDate         pgtype.Timestamptz
}

type MasterStudyPlans []*MasterStudyPlan

func (u *MasterStudyPlans) Add() database.Entity {
	e := &MasterStudyPlan{}
	*u = append(*u, e)
	return e
}

func (m *MasterStudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"study_plan_id",
		"learning_material_id",
		"status",
		"start_date",
		"end_date",
		"available_from",
		"available_to",
		"school_date",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&m.StudyPlanID,
		&m.LearningMaterialID,
		&m.Status,
		&m.StartDate,
		&m.EndDate,
		&m.AvailableFrom,
		&m.AvailableTo,
		&m.SchoolDate,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.DeletedAt,
	}
	return
}

func (m *MasterStudyPlan) TableName() string {
	return "master_study_plan"
}

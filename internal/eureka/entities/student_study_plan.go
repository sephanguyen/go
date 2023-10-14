package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type StudentStudyPlan struct {
	BaseEntity
	StudentID         pgtype.Text
	StudyPlanID       pgtype.Text
	MasterStudyPlanID pgtype.Text
}

func (rcv *StudentStudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_id", "study_plan_id", "created_at", "updated_at", "deleted_at", "master_study_plan_id"}
	values = []interface{}{&rcv.StudentID, &rcv.StudyPlanID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt, &rcv.MasterStudyPlanID}
	return
}

func (rcv *StudentStudyPlan) TableName() string {
	return "student_study_plans"
}

type StudentStudyPlans []*StudentStudyPlan

func (rcv *StudentStudyPlans) Add() database.Entity {
	e := &StudentStudyPlan{}
	*rcv = append(*rcv, e)

	return e
}

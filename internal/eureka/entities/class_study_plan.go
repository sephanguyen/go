package entities

import "github.com/jackc/pgtype"

type ClassStudyPlan struct {
	BaseEntity
	ClassID     pgtype.Int4
	StudyPlanID pgtype.Text
}

func (rcv *ClassStudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "study_plan_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.ClassID, &rcv.StudyPlanID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *ClassStudyPlan) TableName() string {
	return "class_study_plans"
}

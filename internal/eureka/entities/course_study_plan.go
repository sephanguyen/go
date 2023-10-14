package entities

import "github.com/jackc/pgtype"

type CourseStudyPlan struct {
	BaseEntity
	CourseID    pgtype.Text
	StudyPlanID pgtype.Text
}

func (rcv *CourseStudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "study_plan_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.CourseID, &rcv.StudyPlanID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *CourseStudyPlan) TableName() string {
	return "course_study_plans"
}

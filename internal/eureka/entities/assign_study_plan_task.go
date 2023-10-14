package entities

import (
	"github.com/jackc/pgtype"
)

type AssignStudyPlanTask struct {
	BaseEntity
	ID           pgtype.Text
	StudyPlanIDs pgtype.TextArray
	Status       pgtype.Text
	CourseID     pgtype.Text
	ErrorDetail  pgtype.Text
}

func (rcv *AssignStudyPlanTask) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "study_plan_ids", "status", "course_id", "error_detail", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.ID, &rcv.StudyPlanIDs, &rcv.Status, &rcv.CourseID, &rcv.ErrorDetail, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *AssignStudyPlanTask) TableName() string {
	return "assign_study_plan_tasks"
}

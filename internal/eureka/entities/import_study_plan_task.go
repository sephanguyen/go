package entities

import "github.com/jackc/pgtype"

type ImportStudyPlanTask struct {
	TaskID      pgtype.Text
	StudyPlanID pgtype.Text
	Status      pgtype.Text
	ErrorDetail pgtype.Text
	ImportedBy  pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

func (t *ImportStudyPlanTask) FieldMap() ([]string, []interface{}) {
	return []string{
			"task_id",
			"study_plan_id",
			"status",
			"error_detail",
			"imported_by",
			"updated_at",
			"created_at",
		}, []interface{}{
			&t.TaskID,
			&t.StudyPlanID,
			&t.Status,
			&t.ErrorDetail,
			&t.ImportedBy,
			&t.UpdatedAt,
			&t.CreatedAt,
		}
}

func (t *ImportStudyPlanTask) TableName() string {
	return "import_study_plan_task"
}

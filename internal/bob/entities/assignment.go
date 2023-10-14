package entities

import "github.com/jackc/pgtype"

const (
	AssignmentTypeClass      = "ASSIGNMENT_TYPE_CLASS"
	AssignmentTypeSelf       = "ASSIGNMENT_TYPE_SELF"
	AssignmentTypeIndividual = "ASSIGNMENT_TYPE_INDIVIDUAL"
)

type Assignment struct {
	AssignmentID            pgtype.Text
	AssignmentType          pgtype.Text
	AssignedBy              pgtype.Text
	TopicID                 pgtype.Text
	PresetStudyPlanID       pgtype.Text
	StartDate               pgtype.Timestamptz
	EndDate                 pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
	PresetStudyPlanWeeklyID pgtype.Text
	ClassID                 pgtype.Int4
}

func (e *Assignment) FieldMap() ([]string, []interface{}) {
	return []string{
			"assignment_id", "assignment_type", "assigned_by", "topic_id", "preset_study_plan_id", "start_date", "end_date", "updated_at", "created_at", "deleted_at", "preset_study_plan_weekly_id", "class_id",
		}, []interface{}{
			&e.AssignmentID, &e.AssignmentType, &e.AssignedBy, &e.TopicID, &e.PresetStudyPlanID, &e.StartDate, &e.EndDate, &e.UpdatedAt, &e.CreatedAt, &e.DeletedAt, &e.PresetStudyPlanWeeklyID, &e.ClassID,
		}
}

func (e *Assignment) TableName() string {
	return "assignments"
}

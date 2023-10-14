package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AllocateMarker struct {
	AllocateMarkerID   pgtype.Text
	TeacherID          pgtype.Text
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	CreatedBy          pgtype.Text
	BaseEntity
}

type AllocateTeacherItem struct {
	TeacherID                pgtype.Text
	TeacherName              pgtype.Text
	NumberAssignedSubmission int32
}

func (t *AllocateMarker) FieldMap() ([]string, []interface{}) {
	return []string{
			"allocate_marker_id",
			"teacher_id",
			"student_id",
			"study_plan_id",
			"learning_material_id",
			"created_by",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&t.AllocateMarkerID,
			&t.TeacherID,
			&t.StudentID,
			&t.StudyPlanID,
			&t.LearningMaterialID,
			&t.CreatedBy,
			&t.UpdatedAt,
			&t.CreatedAt,
			&t.DeletedAt,
		}
}

func (t *AllocateMarker) TableName() string {
	return "allocate_marker"
}

type AllocateMarkers []*AllocateMarker

func (s *AllocateMarkers) Add() database.Entity {
	e := &AllocateMarker{}
	*s = append(*s, e)

	return e
}

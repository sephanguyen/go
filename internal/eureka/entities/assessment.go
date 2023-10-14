package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Assessment struct {
	BaseEntity
	ID                 pgtype.Text
	CourseID           pgtype.Text
	LearningMaterialID pgtype.Text
}

func (e *Assessment) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"learning_material_id",
			"course_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.ID,
			&e.LearningMaterialID,
			&e.CourseID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *Assessment) TableName() string {
	return "assessment"
}

type Assessments []*Assessment

func (u *Assessments) Add() database.Entity {
	e := &Assessment{}
	*u = append(*u, e)
	return e
}

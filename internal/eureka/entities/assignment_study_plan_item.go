package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AssignmentStudyPlanItem struct {
	BaseEntity
	AssignmentID    pgtype.Text
	StudyPlanItemID pgtype.Text
}

func (rcv *AssignmentStudyPlanItem) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"assignment_id", "study_plan_item_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.AssignmentID, &rcv.StudyPlanItemID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *AssignmentStudyPlanItem) TableName() string {
	return "assignment_study_plan_items"
}

type AssignmentStudyPlanItems []*AssignmentStudyPlanItem

func (rcv *AssignmentStudyPlanItems) Add() database.Entity {
	e := &AssignmentStudyPlanItem{}
	*rcv = append(*rcv, e)

	return e
}

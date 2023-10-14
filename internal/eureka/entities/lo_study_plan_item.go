package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LoStudyPlanItem struct {
	BaseEntity
	StudyPlanItemID pgtype.Text
	LoID            pgtype.Text
}

func (rcv *LoStudyPlanItem) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"lo_id", "study_plan_item_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.LoID, &rcv.StudyPlanItemID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *LoStudyPlanItem) TableName() string {
	return "lo_study_plan_items"
}

type LoStudyPlanItems []*LoStudyPlanItem

func (rcv *LoStudyPlanItems) Add() database.Entity {
	e := &LoStudyPlanItem{}
	*rcv = append(*rcv, e)

	return e
}

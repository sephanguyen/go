package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudyPlanItemDto struct {
	StudyPlanItemID pgtype.Text
	StudyPlanID     pgtype.Text
	LmListID        pgtype.Text
	Name            pgtype.Text
	StartDate       pgtype.Timestamp
	EndDate         pgtype.Timestamp
	DisplayOrder    pgtype.Int4
	Status          pgtype.Text
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
	DeletedAt       pgtype.Timestamp
}

func (s StudyPlanItemDto) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"study_plan_item_id", "study_plan_id", "lm_list_id", "name", "start_date", "end_date", "display_order", "status", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&s.StudyPlanItemID, &s.StudyPlanID, &s.LmListID, &s.Name, &s.StartDate, &s.EndDate, &s.DisplayOrder, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt}
	return
}

func (s StudyPlanItemDto) TableName() string {
	return "lms_study_plan_items"
}

type StudyPlanItemDtos []*StudyPlanItemDto

func (s *StudyPlanItemDtos) Add() database.Entity {
	e := &StudyPlanItemDto{}
	*s = append(*s, e)
	return e
}

func (s StudyPlanItemDto) FromEntity(studyPlanItem domain.StudyPlanItem) (*StudyPlanItemDto, error) {
	e := &StudyPlanItemDto{}
	err := multierr.Combine(
		e.StudyPlanItemID.Set(studyPlanItem.StudyPlanItemID),
		e.StudyPlanID.Set(studyPlanItem.StudyPlanID),
		e.Name.Set(studyPlanItem.Name),
		e.LmListID.Set(studyPlanItem.LmListID),
		e.StartDate.Set(studyPlanItem.StartDate),
		e.EndDate.Set(studyPlanItem.EndDate),
		e.DisplayOrder.Set(studyPlanItem.DisplayOrder),
		e.Status.Set(studyPlanItem.Status),
		e.CreatedAt.Set(studyPlanItem.CreatedAt),
		e.UpdatedAt.Set(studyPlanItem.UpdatedAt),
		e.DeletedAt.Set(studyPlanItem.DeletedAt),
	)
	return e, err
}

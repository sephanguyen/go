package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LmListDto struct {
	LmListID  pgtype.Text
	LmIDs     pgtype.TextArray
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	DeletedAt pgtype.Timestamp
}

func (s LmListDto) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"lm_list_id", "lm_ids", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&s.LmListID, &s.LmIDs, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt}
	return
}

func (s LmListDto) TableName() string {
	return "lms_learning_material_list"
}

type LmListDtos []*LmListDto

func (u *LmListDtos) Add() database.Entity {
	e := &LmListDto{}
	*u = append(*u, e)
	return e
}

func (s LmListDto) ToLmListEntity(lmList domain.LmList) (*LmListDto, error) {
	e := &LmListDto{}
	err := multierr.Combine(
		e.LmListID.Set(lmList.LmListID),
		e.LmIDs.Set(lmList.LmIDs),
		e.CreatedAt.Set(lmList.CreatedAt),
		e.UpdatedAt.Set(lmList.UpdatedAt),
		e.DeletedAt.Set(lmList.DeletedAt),
	)
	return e, err
}

package postgres

import (
	"context"
	"time"

	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type StudyPlanItemRepo struct {
	DB database.Ext
}

func (repo *StudyPlanItemRepo) UpsertStudyPlanItems(ctx context.Context, studyPlanItems []*dto.StudyPlanItemDto) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.Upsert")
	defer span.End()

	query := `INSERT INTO %s (%s) 
	VALUES %s ON CONFLICT ON CONSTRAINT lms_study_plan_items_pkey
	DO UPDATE SET 
		name = excluded.name,
		start_date = excluded.start_date, 
		end_date = excluded.end_date,
		updated_at = excluded.updated_at,
		deleted_at = excluded.deleted_at,
		display_order = excluded.display_order`

	now := time.Now()
	for _, studyPlanItem := range studyPlanItems {
		err := multierr.Combine(
			studyPlanItem.CreatedAt.Set(now),
			studyPlanItem.UpdatedAt.Set(now),
			studyPlanItem.DeletedAt.Set(nil),
		)
		if err != nil {
			return errors.NewConversionError("StudyPlanItemRepo.Upsert", err)
		}
	}
	err := eureka_db.BulkUpsert(ctx, repo.DB, query, studyPlanItems)
	if err != nil {
		return errors.NewDBError("eureka_db.StudyPlanItemRepo", err)
	}
	return nil
}

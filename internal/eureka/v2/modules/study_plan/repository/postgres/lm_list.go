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

type LmListRepo struct {
	DB database.Ext
}

func (repo *LmListRepo) UpsertLearningMaterialsIDList(ctx context.Context, lmLists []*dto.LmListDto) error {
	ctx, span := interceptors.StartSpan(ctx, "LmListRepo.UpsertLearningMaterialsIDList")
	defer span.End()

	query := `INSERT INTO %s (%s) 
	VALUES %s ON CONFLICT ON CONSTRAINT lms_learning_material_list_pkey
	DO UPDATE SET 
		lm_ids = excluded.lm_ids,
		updated_at = excluded.updated_at,
		deleted_at = excluded.deleted_at
	`

	now := time.Now()
	for _, lmList := range lmLists {
		err := multierr.Combine(
			lmList.CreatedAt.Set(now),
			lmList.UpdatedAt.Set(now),
			lmList.DeletedAt.Set(nil),
		)
		if err != nil {
			return errors.NewConversionError("LmListRepo.UpsertLearningMaterialsIDList", err)
		}
	}
	err := eureka_db.BulkUpsert(ctx, repo.DB, query, lmLists)
	if err != nil {
		return errors.NewDBError("LmListRepo.BulkUpsert", err)
	}
	return nil
}

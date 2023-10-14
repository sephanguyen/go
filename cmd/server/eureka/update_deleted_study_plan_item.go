package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_update_deleted_study_plan_item", RunUpdateDeletedStudyPlanItem)
}

func RunUpdateDeletedStudyPlanItem(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	eurekaDBConn := rsc.DB().DB.(*pgxpool.Pool)

	migrateLoDeletedStudyPlanItem(ctx, eurekaDBConn)
	zapLogger.Info("Completed update deleted study plan item type learning objective")
	migrateAssDeletedStudyPlanItem(ctx, eurekaDBConn)
	zapLogger.Info("Completed update deleted study plan item type assignment")
	return nil
}

func migrateLoDeletedStudyPlanItem(ctx context.Context, db *pgxpool.Pool) {
	stmt := `UPDATE study_plan_items spi SET deleted_at = NULL, status = 'STUDY_PLAN_ITEM_STATUS_ARCHIVED', updated_at = '2022-08-17 00:00:00.000 +0700'
	where spi.study_plan_item_id = ANY(select spi.study_plan_item_id from study_plan_items spi join learning_objectives lo 
	on lo.lo_id = spi.content_structure ->> 'lo_id' where lo.deleted_at is null and spi.deleted_at is not null limit 100)`
	for {
		rows, err := db.Exec(ctx, stmt)
		if err != nil {
			zapLogger.Fatal("update deleted study plan item type learning objective err:", zap.Error(err))
			return
		}
		if rows.RowsAffected() == 0 {
			break
		}
	}
}

func migrateAssDeletedStudyPlanItem(ctx context.Context, db *pgxpool.Pool) {
	stmt := `UPDATE study_plan_items spi SET deleted_at = NULL, status = 'STUDY_PLAN_ITEM_STATUS_ARCHIVED', updated_at = '2022-08-17 00:00:00.000 +0700'
	where spi.study_plan_item_id = ANY(select spi.study_plan_item_id from study_plan_items spi join assignments ass 
		on ass.assignment_id = spi.content_structure ->> 'assignment_id' where ass.deleted_at is null and spi.deleted_at is not null limit 100)`

	for {
		rows, err := db.Exec(ctx, stmt)
		if err != nil {
			zapLogger.Fatal("update deleted study plan item type assignment err:", zap.Error(err))
			return
		}
		if rows.RowsAffected() == 0 {
			break
		}
	}
}

package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_migrate_master_study_plan", RunMigrateMasterStudyPlans)
}

// RunMigrateMasterStudyPlans
func RunMigrateMasterStudyPlans(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	eurekaDBConn := rsc.DB()

	now := time.Now()

	if err := database.ExecInTx(ctx, eurekaDBConn, func(ctx context.Context, tx pgx.Tx) error {
		stmt := `
		INSERT INTO
		public.master_study_plan (
			study_plan_id,
			learning_material_id,
			status,
			start_date,
			end_date,
			available_from,
			available_to,
			created_at,
			updated_at,
			deleted_at,
			school_date,
			resource_path
		)
		SELECT
			sp.study_plan_id,
			coalesce(NULLIF(spi.content_structure ->> 'lo_id',''),spi.content_structure->>'assignment_id',''),
			spi.status,
			spi.start_date,
			spi.end_date,
			spi.available_from,
			spi.available_to,
			spi.created_at,
			spi.updated_at,
			spi.deleted_at,
			spi.school_date,
			spi.resource_path    
		FROM public.study_plans sp JOIN public.study_plan_items spi ON sp.study_plan_id = spi.study_plan_id
		WHERE sp.master_study_plan_id IS NULL AND sp.study_plan_type='STUDY_PLAN_TYPE_COURSE' AND sp.deleted_at IS NULL 
		AND (spi.start_date IS NOT NULL 
			OR spi.end_date IS NOT NULL 
			OR spi.available_from IS NOT NULL
			OR spi.available_to IS NOT NULL
			OR spi.school_date IS NOT NULL)
		ON CONFLICT ON constraint learning_material_id_study_plan_id_pk DO nothing
		`
		_, err := tx.Exec(ctx, stmt)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		zapLogger.Error("migrate master study plan err", zap.Error(err))
	}

	zapLogger.Info(fmt.Sprintf("migrate master study plan completed and it took %vms", time.Since(now).Milliseconds()))
	return nil
}

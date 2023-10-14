package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_migrate_study_plan_items_to_individual_study_plan", RunMigrateStudyPlanItemsToIndividualStudyPlan)
}

// RunMigrateStudyPlanItemsToIndividualStudyPlan
func RunMigrateStudyPlanItemsToIndividualStudyPlan(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	eurekaDBConn := rsc.DB().DB.(*pgxpool.Pool)

	err := migrateStudyPlanItemsToIndividualStudyPlan(ctx, eurekaDBConn)
	if err != nil {
		zapLogger.Error("migrateStudyPlanItemsToIndividualStudyPlan", zap.Error(err))
	}
	zapLogger.Info("Complete migrate study plan items to individual study plan")
	return nil
}

func migrateStudyPlanItemsToIndividualStudyPlan(ctx context.Context, db *pgxpool.Pool) error {
	migrateQuery := `with individual as (
    select spi.*, sp.study_plan_type, ssp.student_id, ssp.master_study_plan_id
    from study_plan_items spi
    join study_plans sp using (study_plan_id)
    join student_study_plans ssp using (study_plan_id)
	left join study_plan_items master_spi on spi.copy_study_plan_item_id = master_spi.study_plan_item_id
    where (
			spi.copy_study_plan_item_id is not null
			and (
				spi.start_date != master_spi.start_date
				or spi.available_from != master_spi.available_from
				or spi.available_to != master_spi.available_to
				or spi.end_date != master_spi.end_date
				or spi.school_date != master_spi.school_date
				or spi.status != master_spi.status
			)
		)
        or sp.study_plan_type = 'STUDY_PLAN_TYPE_INDIVIDUAL'
	)
	insert into individual_study_plan (
		study_plan_id,
		learning_material_id,
		student_id,
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
	select 
		COALESCE(master_study_plan_id,study_plan_id), 
		coalesce(NULLIF(content_structure ->> 'lo_id',''),content_structure->>'assignment_id', ''),
		student_id,
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
    from individual
	on conflict on constraint learning_material_id_student_id_study_plan_id_pk
	do nothing;`
	start := time.Now()
	err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		_, err := db.Exec(ctx, migrateQuery)
		return err
	})
	zapLogger.Info(fmt.Sprintf("time to complete migrate: %s", time.Since(start)))
	if err != nil {
		return fmt.Errorf("failed to migrate study plan items to individual study plan: %w", err)
	}
	return nil
}

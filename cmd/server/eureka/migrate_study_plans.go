package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_migrate_study_plan", RunMigrateStudyPlans)
}

// RunMigrateStudyPlans
func RunMigrateStudyPlans(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	eurekaDBConn := rsc.DB().DB.(*pgxpool.Pool)

	err := migrateStudyPlanItems(ctx, eurekaDBConn)
	if err != nil {
		zapLogger.Error("migrateStudyPlanItems", zap.Error(err))
	}
	zapLogger.Info("Complete migrate study plan")
	err = migrateAssignmentStudyPlanItems(ctx, eurekaDBConn)
	if err != nil {
		zapLogger.Error("migrateAssignmentStudyPlanItems", zap.Error(err))
	}
	err = migrateLoStudyPlanItems(ctx, eurekaDBConn)
	if err != nil {
		zapLogger.Error("migrateLoStudyPlanItems", zap.Error(err))
	}
	zapLogger.Info("Complete migrate all table for study plan")
	return nil
}

func migrateStudyPlanItems(ctx context.Context, db *pgxpool.Pool) error {
	var currentStudyPlanItemID pgtype.Text
	firstStudyPlanItems := `WITH sp AS(
		SELECT study_plan_item_id, school_id
		FROM study_plan_items spi join study_plans sp  on spi.study_plan_id = sp.study_plan_id
		WHERE sp.school_id IS NOT NULL
			AND (spi.resource_path IS NULL OR LENGTH(spi.resource_path)=0)
		ORDER BY study_plan_item_id asc
		LIMIT 100
	)
	UPDATE study_plan_items spi SET resource_path = sp.school_id::text
	FROM sp
	WHERE spi.study_plan_item_id = sp.study_plan_item_id
	RETURNING sp.study_plan_item_id;`

	rows, err := db.Query(ctx, firstStudyPlanItems)
	if err != nil {
		return fmt.Errorf("failed to find first study plan item: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to scan study plan item: %w", err)
		}
	}

	var prevStudyPlanItemID pgtype.Text
	for currentStudyPlanItemID.Status == pgtype.Present && prevStudyPlanItemID.String != currentStudyPlanItemID.String {
		prevStudyPlanItemID = currentStudyPlanItemID
		updateQuery := `WITH sp AS(
			SELECT study_plan_item_id, school_id
			FROM study_plan_items spi join study_plans sp  on spi.study_plan_id = sp.study_plan_id
			WHERE study_plan_item_id >$1
				AND sp.school_id IS NOT NULL
				AND (spi.resource_path IS NULL OR LENGTH(spi.resource_path)=0)
			ORDER BY study_plan_item_id asc
			LIMIT 100
		)
		UPDATE study_plan_items spi SET resource_path = sp.school_id::text
		FROM sp
		WHERE spi.study_plan_item_id = sp.study_plan_item_id
		RETURNING sp.study_plan_item_id;`
		rows, err = db.Query(ctx, updateQuery, &currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to update study plan item: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&currentStudyPlanItemID)
			if err != nil {
				return fmt.Errorf("failed to scan study plan item: %w", err)
			}
		}
		zapLogger.Info("migrateStudyPlanItems", zap.String("currentStudyPlanItemID", currentStudyPlanItemID.String))
	}
	return nil
}

func migrateAssignmentStudyPlanItems(ctx context.Context, db *pgxpool.Pool) error {
	var currentStudyPlanItemID pgtype.Text
	findFirstStudyPlanItemID := `WITH sp AS(
		SELECT aspi.study_plan_item_id, spi.resource_path 
		FROM assignment_study_plan_items aspi 
			 join study_plan_items spi on aspi.study_plan_item_id = spi.study_plan_item_id 
		WHERE aspi.resource_path IS NULL OR LENGTH(aspi.resource_path)=0
		ORDER BY aspi.study_plan_item_id asc
		LIMIT 100
	) 
	UPDATE assignment_study_plan_items aspi SET resource_path = sp.resource_path::text
	FROM sp
	WHERE aspi.study_plan_item_id = sp.study_plan_item_id
	RETURNING sp.study_plan_item_id;`

	rows, err := db.Query(ctx, findFirstStudyPlanItemID)
	if err != nil {
		return fmt.Errorf("failed to find first assignment study plan item: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to scan assignment study plan item: %w", err)
		}
	}

	var prevStudyPlanItemID pgtype.Text
	for currentStudyPlanItemID.Status == pgtype.Present && prevStudyPlanItemID.String != currentStudyPlanItemID.String {
		prevStudyPlanItemID = currentStudyPlanItemID
		updateQuery := `WITH sp AS(
			SELECT aspi.study_plan_item_id, spi.resource_path 
			FROM assignment_study_plan_items aspi 
				 join study_plan_items spi on aspi.study_plan_item_id = spi.study_plan_item_id 
			WHERE aspi.study_plan_item_id > $1
				AND (aspi.resource_path IS NULL OR LENGTH(aspi.resource_path)=0)
			ORDER BY aspi.study_plan_item_id asc
			LIMIT 100
		) 
		UPDATE assignment_study_plan_items aspi SET resource_path = sp.resource_path::text
		FROM sp
		WHERE aspi.study_plan_item_id = sp.study_plan_item_id
		RETURNING sp.study_plan_item_id;`
		rows, err = db.Query(ctx, updateQuery, &currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to update assignment study plan item: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&currentStudyPlanItemID)
			if err != nil {
				return fmt.Errorf("failed to scan assignment study plan item: %w", err)
			}
		}
		zapLogger.Info("migrateAssignmentStudyPlanItems", zap.String("currentAssignmentStudyPlanItemID", currentStudyPlanItemID.String))
	}
	return nil
}

func migrateLoStudyPlanItems(ctx context.Context, db *pgxpool.Pool) error {
	var currentStudyPlanItemID pgtype.Text
	findFirstStudyPlanItemID := `WITH sp AS(
		SELECT lspi.study_plan_item_id, spi.resource_path 
		FROM lo_study_plan_items lspi 
			 join study_plan_items spi on lspi.study_plan_item_id = spi.study_plan_item_id 
		WHERE (lspi.resource_path IS NULL OR LENGTH(lspi.resource_path)=0)
		ORDER BY lspi.study_plan_item_id asc
		LIMIT 100
	) 
	UPDATE lo_study_plan_items lspi SET resource_path = sp.resource_path::text
	FROM sp
	WHERE lspi.study_plan_item_id = sp.study_plan_item_id
	RETURNING sp.study_plan_item_id;`

	rows, err := db.Query(ctx, findFirstStudyPlanItemID)
	if err != nil {
		return fmt.Errorf("failed to find first assignment study plan item: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to scan assignment study plan item: %w", err)
		}
	}

	var prevStudyPlanItemID pgtype.Text
	for currentStudyPlanItemID.Status == pgtype.Present && prevStudyPlanItemID.String != currentStudyPlanItemID.String {
		prevStudyPlanItemID = currentStudyPlanItemID
		updateQuery := `WITH sp AS(
			SELECT lspi.study_plan_item_id, spi.resource_path 
			FROM lo_study_plan_items lspi 
				 join study_plan_items spi on lspi.study_plan_item_id = spi.study_plan_item_id 
			WHERE lspi.study_plan_item_id > $1
				AND (lspi.resource_path IS NULL OR LENGTH(lspi.resource_path)=0)
			ORDER BY lspi.study_plan_item_id asc
			LIMIT 100
		) 
		UPDATE lo_study_plan_items lspi SET resource_path = sp.resource_path::text
		FROM sp
		WHERE lspi.study_plan_item_id = sp.study_plan_item_id
		RETURNING sp.study_plan_item_id;`
		rows, err = db.Query(ctx, updateQuery, &currentStudyPlanItemID)
		if err != nil {
			return fmt.Errorf("failed to update lo study plan item: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&currentStudyPlanItemID)
			if err != nil {
				return fmt.Errorf("failed to scan lo study plan item: %w", err)
			}
		}
		zapLogger.Info("migrateLoStudyPlanItems", zap.String("currentLoStudyPlanItemID", currentStudyPlanItemID.String))
	}
	return nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ClassStudyPlanRepo struct {
}

const bulkUpsertClassStudyPlan = `INSERT INTO %s (%s) 
VALUES (%s) 
ON CONFLICT ON CONSTRAINT class_study_plans_pk DO UPDATE 
SET
	updated_at = excluded.updated_at`

func (r *ClassStudyPlanRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, classStudyPlans []*entities.ClassStudyPlan) error {
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertClassStudyPlan, classStudyPlans)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertClassStudyPlan error: %s", err.Error())
	}
	return nil
}

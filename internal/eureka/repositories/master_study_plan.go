package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type MasterStudyPlanRepo struct {
}

const masterStudyPlanBulkInsertStmtTpl = `INSERT INTO %s (%s)
VALUES %s
ON CONFLICT ON CONSTRAINT learning_material_id_study_plan_id_pk DO UPDATE 
SET
status = excluded.status,
start_date = excluded.start_date,
end_date = excluded.end_date,
available_from = excluded.available_from,
available_to = excluded.available_to,
school_date = excluded.school_date,
updated_at = NOW();`

func (m *MasterStudyPlanRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.MasterStudyPlan) error {
	err := dbeureka.BulkUpsert(ctx, db, masterStudyPlanBulkInsertStmtTpl, items)
	if err != nil {
		return fmt.Errorf("database.BulkUpsert error: %s", err.Error())
	}
	return nil
}

func (m *MasterStudyPlanRepo) BulkUpdateTime(ctx context.Context, db database.QueryExecer, items []*entities.MasterStudyPlan) error {
	queueFn := func(b *pgx.Batch, e *entities.MasterStudyPlan) {
		query := `UPDATE master_study_plan 
		SET start_date = $3, end_date = $4, available_from = $5, available_to = $6, updated_at = now()
		WHERE study_plan_id = $1 and learning_material_id = $2`
		b.Queue(query, &e.StudyPlanID, &e.LearningMaterialID, &e.StartDate, &e.EndDate, &e.AvailableFrom, &e.AvailableTo)
	}

	b := &pgx.Batch{}
	for _, each := range items {
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (m *MasterStudyPlanRepo) FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.MasterStudyPlan, error) {
	masterStudyPlans := []*entities.MasterStudyPlan{}
	query := `SELECT study_plan_id, learning_material_id FROM study_plan_tree WHERE study_plan_id = $1`

	rows, err := db.Query(ctx, query, &studyPlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		masterStudyPlan := &entities.MasterStudyPlan{}
		if err := rows.Scan(&masterStudyPlan.StudyPlanID, &masterStudyPlan.LearningMaterialID); err != nil {
			return nil, err
		}

		masterStudyPlans = append(masterStudyPlans, masterStudyPlan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return masterStudyPlans, nil
}

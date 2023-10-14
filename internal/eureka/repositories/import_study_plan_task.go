package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ImportStudyPlanTaskRepo struct {
}

func (s *ImportStudyPlanTaskRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.ImportStudyPlanTask) error {
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (s *ImportStudyPlanTaskRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.ImportStudyPlanTask) error {
	if _, err := database.UpdateFields(ctx, e, db.Exec, "task_id", []string{"status", "updated_at", "error_detail"}); err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}
	return nil
}

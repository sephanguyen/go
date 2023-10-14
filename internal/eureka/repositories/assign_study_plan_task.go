package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AssignStudyPlanTaskRepo struct {
}

func (r *AssignStudyPlanTaskRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.AssignStudyPlanTask) (pgtype.Text, error) {
	now := time.Now()

	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)

	var id pgtype.Text
	err := database.InsertReturning(ctx, e, db, "id", &id)
	return id, err
}

func (r *AssignStudyPlanTaskRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, id pgtype.Text, status pgtype.Text) error {
	now := time.Now()
	e := &entities.AssignStudyPlanTask{}
	e.ID.Set(id)
	e.Status.Set(status)
	e.UpdatedAt.Set(now)
	_, err := database.UpdateFields(ctx, e, db.Exec, "id", []string{"status", "updated_at"})
	if err != nil {
		return err
	}
	return nil
}

type AssignStudyPlanTaskDetailErrorArgs struct {
	ID          pgtype.Text
	Status      pgtype.Text
	ErrorDetail pgtype.Text
}

func (r *AssignStudyPlanTaskRepo) UpdateDetailError(ctx context.Context, db database.QueryExecer, errDetail *AssignStudyPlanTaskDetailErrorArgs) error {
	e := &entities.AssignStudyPlanTask{}
	updateErrorStmt := `UPDATE %s SET status = $1, error_detail = $2, updated_at = now() WHERE id = $3`
	_, err := db.Exec(ctx, fmt.Sprintf(updateErrorStmt, e.TableName()), errDetail.Status, errDetail.ErrorDetail, errDetail.ID)
	if err != nil {
		return err
	}
	return nil
}

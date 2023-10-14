package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type StudentPackageLogRepo struct{}

func (r *StudentPackageLogRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentPackageLog) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageLogRepo.Create")
	defer span.End()

	if err := e.CreatedAt.Set(time.Now()); err != nil {
		return fmt.Errorf("CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"student_package_log_id", "resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentPackageLog: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentPackageLog: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

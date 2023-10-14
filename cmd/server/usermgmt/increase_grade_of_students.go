package usermgmt

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func RunIncreaseGradeOfStudents(ctx context.Context, c *configurations.Config, createdAt, organizationID string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----START: Increasing grade of students job-----")
	defer zLogger.Sugar().Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	stmt := `
		UPDATE public.students 
		SET previous_grade = current_grade, current_grade = 
			CASE current_grade
				WHEN 0 THEN 0 
				WHEN 16 THEN 16 
				ELSE current_grade + 1
			END
		WHERE deleted_at IS NULL 
		AND created_at::date <= $1
		AND resource_path = $2;
	`
	err = database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, stmt, createdAt, organizationID)
		return err
	})
	if err != nil {
		zLogger.Sugar().Fatalf("RunIncreaseGradeOfStudents err: %v", err)
	}

	zLogger.Sugar().Info("-----DONE: Increasing grade of students job-----")
}

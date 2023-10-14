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

	"go.uber.org/zap"
)

var updateEnrollmentQuery = `UPDATE public.students
	SET enrollment_status = $1
	WHERE enrollment_status = $2
	AND resource_path = $3`

func RunMigrateStudentEnrollmentOriginalStatus(ctx context.Context, c *configurations.Config, newStatus, originStatus, resourcePath string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration update original status to new status-----")
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

	orgQuery := "SELECT organization_id, name FROM organizations WHERE resource_path = $1"
	organizations, err := dbPool.Query(ctx, orgQuery, resourcePath)
	if err != nil {
		zLogger.Fatal("Get orgs failed")
	}
	defer organizations.Close()
	for organizations.Next() {
		var organizationID, name string

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zLogger.Sugar().Infof("failed to scan an orgs row: %s", err)
			continue
		}

		ctx = auth.InjectFakeJwtToken(ctx, organizationID)
		result, err := dbPool.Exec(ctx, updateEnrollmentQuery, newStatus, originStatus, resourcePath)
		if err != nil {
			zLogger.Sugar().Infof("update student_enrollment_status failed")
		}
		zLogger.Sugar().Infof("Done migration for update original status to new status. Result migrate is : %v", result)
	}
}

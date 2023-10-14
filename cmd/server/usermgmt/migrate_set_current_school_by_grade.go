package usermgmt

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

func RunMigrateSetCurrentSchoolByGrade(ctx context.Context, c *configurations.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration set current school by grade-----")
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

	orgQuery := "SELECT organization_id, name FROM organizations"
	organizations, err := dbPool.Query(ctx, orgQuery)
	if err != nil {
		zLogger.Fatal("Get orgs failed")
	}
	defer organizations.Close()
	for organizations.Next() {
		var organizationID, name pgtype.Text

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zLogger.Sugar().Infof("failed to scan an orgs row: %s", err)
			continue
		}

		ctx = auth.InjectFakeJwtToken(ctx, organizationID.String)
		err = migrateSetCurrentSchoolOrg(ctx, dbPool, organizationID.String)
		if err != nil {
			zLogger.Sugar().Errorf("set current school org: %s has error :%s", organizationID.String, err.Error())
			break
		}
		zLogger.Sugar().Infof("Done migration for %s. Migrate success", name)
	}
}

func migrateSetCurrentSchoolOrg(ctx context.Context, db database.QueryExecer, organizationID string) error {
	if err := (&repository.SchoolHistoryRepo{}).SetCurrentSchool(ctx, db, database.Text(organizationID)); err != nil {
		return errorx.ToStatusError(err)
	}
	return nil
}

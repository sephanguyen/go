package usermgmt

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func RunMigrateStudentFullNameToLastNameAndFirstName(ctx context.Context, bobCfg *configurations.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	defer stop()

	zLogger = logger.NewZapLogger("debug", bobCfg.Common.Environment == "local")
	zLogger.Sugar().Info("-----START: Migrate student full name to last name and first name-----")
	defer zLogger.Sugar().Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, bobCfg.PostgresV2.Databases["bob"])
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
		zLogger.Fatal("Get organizations failed")
	}
	defer organizations.Close()
	for organizations.Next() {
		var organizationID, name pgtype.Text

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zLogger.Sugar().Infof("failed to scan an orgs row: %s", err)
			continue
		}
		stmt := `
			UPDATE users AS u
			SET last_name = split_part(u.name, ' ', 1),
    			first_name  = CASE
       			WHEN array_length(regexp_split_to_array(u.name, '\s'), 1) > 1 THEN (SELECT regexp_replace(u.name, '.*?\s', '')) ELSE '' END
			WHERE u.first_name = '' and u.name != '';
		`
		ctx = auth.InjectFakeJwtToken(ctx, organizationID.String)
		err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, stmt)
			return err
		})
		if err != nil {
			zLogger.Sugar().Fatalf("RunMigrateStudentFullNameToLastNameAndFirstName err: %v", err)
		}

		zLogger.Sugar().Info("-----DONE: Migrate student full name to last name and first name-----")
	}
}

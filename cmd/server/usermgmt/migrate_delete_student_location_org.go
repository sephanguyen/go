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
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var (
	zLogger *zap.Logger

	studentIDsScanQuery = `SELECT uap.user_id
	FROM user_access_paths uap
	INNER JOIN locations l ON uap.location_id = l.location_id 
	INNER JOIN location_types lt ON l.location_type = lt.location_type_id 
	WHERE lt.name = $1 AND uap.deleted_at IS NULL
	AND uap.resource_path = $2
	LIMIT $3;
	`
)

func RunMigrateDeleteStudentLocationOrg(ctx context.Context, c *configurations.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration delete student_location org-----")
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
		migrated := migrateDeleteStudentLocationOrg(ctx, dbPool, organizationID.String)
		zLogger.Sugar().Infof("Done migration for %s. Total record migrate is : %v", name, migrated)
	}
}

func migrateDeleteStudentLocationOrg(ctx context.Context, dbPool *pgxpool.Pool, organizationID string) int {
	perBatch := 1000
	studentsMigrated := 0

	for {
		listStudentIDs := getListStudentIDs(ctx, dbPool, organizationID, perBatch)
		if len(listStudentIDs) == 0 {
			break
		}

		if err := deleteStudentLocationOrg(ctx, dbPool, listStudentIDs); err != nil {
			zLogger.Sugar().Errorf("delete student_location org: %s has error :%s", listStudentIDs, err.Error())
			break
		}

		studentsMigrated += len(listStudentIDs)
	}

	return studentsMigrated
}

func getListStudentIDs(ctx context.Context, db database.QueryExecer, organizationID string, limit int) []string {
	studentRows, err := db.Query(ctx, studentIDsScanQuery, domain.DefaultLocationType, organizationID, limit)
	if err != nil {
		zLogger.Fatal(err.Error())
	}
	defer studentRows.Close()
	if studentRows.Err() != nil {
		zLogger.Fatal(err.Error())
	}
	studentIDs := []string{}
	for studentRows.Next() {
		studentID := ""

		if err = studentRows.Scan(&studentID); err != nil {
			zLogger.Sugar().Errorf("failed to scan studentID in user_access_paths: %s", err.Error())
			continue
		}

		studentIDs = append(studentIDs, studentID)
	}
	return studentIDs
}

func deleteStudentLocationOrg(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	if err := (&repository.UserAccessPathRepo{}).Delete(ctx, db, database.TextArray(studentIDs)); err != nil {
		return errorx.ToStatusError(err)
	}

	return nil
}

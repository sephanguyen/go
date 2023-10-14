package fatima

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	bob_cfg "github.com/manabie-com/backend/internal/bob/configurations"
	fatima_cfg "github.com/manabie-com/backend/internal/fatima/configurations"
	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func RunMigrateStudentPackagesToStudentPackageAccessPath(ctx context.Context, fatimaCfg *fatima_cfg.Config, bobCfg *bob_cfg.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", bobCfg.Common.Environment == "local")
	zLogger.Sugar().Info("-----START: Migrating student_packages to student_package_access_path job-----")
	defer zLogger.Sugar().Sync()

	dbconf, exists := fatimaCfg.PostgresV2.Databases["fatima"]
	if !exists {
		panic("config for fatima db does not exist")
	}
	fatimaDB, fatimaDBCancel, err := database.NewPool(ctx, zLogger, dbconf)
	if err != nil {
		panic(fmt.Errorf("failed to connect to fatima db: %s", err))
	}
	defer func() {
		if err := fatimaDBCancel(); err != nil {
			zLogger.Error("fatimaDBCancel failed", zap.Error(err))
		}
	}()

	bobDB, bobDBCancel, err := database.NewPool(ctx, zLogger, bobCfg.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := bobDBCancel(); err != nil {
			zLogger.Error("bobDBCancel() failed", zap.Error(err))
		}
	}()

	orgQuery := "SELECT organization_id, name FROM organizations"
	organizations, err := bobDB.Query(ctx, orgQuery)
	if err != nil {
		zLogger.Fatal("Get orgs failed")
	}
	defer organizations.Close()

	for organizations.Next() {
		var organizationID, name string

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zLogger.Sugar().Errorf("failed to scan an orgs row: %s", err)
			continue
		}

		ctx = auth.InjectFakeJwtToken(ctx, organizationID)

		perBatch := 100
		var latestID string
		var totalStudentPackageAccessPath int
		for {
			studentPackages, err := getStudentPackagesByResourcePath(ctx, fatimaDB, organizationID, latestID, perBatch)
			if err != nil {
				zLogger.Sugar().Errorf("failed to getStudentPackagesByResourcePath: %s, resource_path: ", err, organizationID)
			}
			if len(studentPackages) == 0 {
				zLogger.Sugar().Infof("Query return 0 rows, done migrate. Total record migrate is : %v", totalStudentPackageAccessPath)
				break
			}
			latestID = studentPackages[len(studentPackages)-1].ID.String

			studentPackagesAccessPaths := []*entities.StudentPackageAccessPath{}
			for _, sp := range studentPackages {
				spap, err := generateListStudentPackageAccessPathsFromStudentPackage(sp)
				if err != nil {
					zLogger.Sugar().Errorf("failed to generateListStudentPackageAccessPathsFromStudentPackage: %s, student_package_id: %s", err, sp.ID.String)
					continue
				}
				studentPackagesAccessPaths = append(studentPackagesAccessPaths, spap...)
			}

			spapRepo := &repositories.StudentPackageAccessPathRepo{}
			err = spapRepo.BulkUpsert(ctx, fatimaDB, studentPackagesAccessPaths)
			if err != nil {
				zLogger.Sugar().Errorf("failed to BulkUpsert: %s", err)
				continue
			}

			totalStudentPackageAccessPath += len(studentPackagesAccessPaths)
		}
	}

	zLogger.Sugar().Info("-----DONE: Migrating student_packages to student_package_access_path job-----")
}

func toStudentPackageAccessPath(sp *entities.StudentPackage, courseID, locationID string) (*entities.StudentPackageAccessPath, error) {
	spap := &entities.StudentPackageAccessPath{}
	database.AllNullEntity(spap)

	if err := multierr.Combine(
		spap.StudentPackageID.Set(sp.ID),
		spap.CourseID.Set(courseID),
		spap.StudentID.Set(sp.StudentID),
		spap.LocationID.Set(locationID),
	); err != nil {
		return nil, err
	}

	return spap, nil
}

func generateListStudentPackageAccessPathsFromStudentPackage(sp *entities.StudentPackage) ([]*entities.StudentPackageAccessPath, error) {
	studentPackageAccessPaths := make([]*entities.StudentPackageAccessPath, 0)

	var locationIDs []string
	if err := sp.LocationIDs.AssignTo(&locationIDs); err != nil {
		return nil, err
	}

	courseIDs, err := sp.GetCourseIDs()
	if err != nil {
		return nil, err
	}

	for _, courseID := range courseIDs {
		if len(locationIDs) > 0 {
			for _, locationID := range locationIDs {
				spap, err := toStudentPackageAccessPath(sp, courseID, locationID)
				if err != nil {
					return nil, err
				}
				studentPackageAccessPaths = append(studentPackageAccessPaths, spap)
			}
		} else {
			spap, err := toStudentPackageAccessPath(sp, courseID, "")
			if err != nil {
				return nil, err
			}
			studentPackageAccessPaths = append(studentPackageAccessPaths, spap)
		}
	}

	return studentPackageAccessPaths, nil
}

func getStudentPackagesByResourcePath(ctx context.Context, db database.QueryExecer, resourcePath, latestID string, limit int) ([]*entities.StudentPackage, error) {
	e := &entities.StudentPackage{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE resource_path = $1 AND student_package_id > $2 ORDER BY student_package_id LIMIT $3`, strings.Join(fields, ", "), e.TableName())

	studentPackages := entities.StudentPackages{}
	err := database.Select(ctx, db, query, resourcePath, latestID, limit).ScanAll(&studentPackages)
	if err != nil {
		return nil, err
	}

	return studentPackages, nil
}

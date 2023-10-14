package fatima

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/fatima/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	bootstrap.RegisterJob("fatima_migrate_student_packages", runMigrateResourcePath)
}

var (
	studentIDsScanQuery = `SELECT student_id
	FROM student_packages
	WHERE (resource_path is null or resource_path = '') and student_id > $1 
	GROUP BY student_id
	ORDER BY student_id limit $2;`
	resourcePathUpdateQuery    = `UPDATE student_packages SET updated_at= now(), resource_path = $1 WHERE student_id = ANY($2);`
	getPackageIDByStudentQuery = `SELECT package_id
	FROM student_packages
	WHERE student_id = ANY($1) AND package_id IS NOT NULL
	GROUP BY package_id;`
	resourcePathUpdateForPackageQuery = `UPDATE packages SET updated_at= now(), resource_path = $1 WHERE package_id = ANY($2);`
)

func migrateResourcePath(
	ctx context.Context,
	l *zap.SugaredLogger,
	fatimaDB *database.DBTrace,
	bobCon *grpc.ClientConn,
) (int64, int64, error) {
	l.Info("Migrate resource path for payment")
	perBatch := 100
	studentReaderClient := bob_pbv1.NewStudentReaderServiceClient(bobCon)
	var totalUpdatedStudentPackage, totalUpdatedPackage int64
	latestID := ""
	for {
		studentIDs, err := getStudentIDsInStudentPackage(ctx, l, fatimaDB, perBatch, latestID)
		if err != nil {
			return 0, 0, fmt.Errorf("getStudentIDsInStudentPackage: %s", err)
		}
		if len(studentIDs) == 0 {
			l.Infof("Query return 0 rows, done migrate. Total record migrate is : %v", totalUpdatedStudentPackage)
			break
		}
		latestID = studentIDs[len(studentIDs)-1]
		res, err := studentReaderClient.GetListSchoolIDsByStudentIDs(ctx, &bob_pbv1.GetListSchoolIDsByStudentIDsRequest{StudentIds: studentIDs})
		if err != nil {
			l.Errorf("Error when get school ids from bob %v", err.Error())
			continue
		}
		if len(res.SchoolIds) == 0 {
			continue
		}
		updatedRowsStudentPackage, updatedRowsPackage, err := updateResourcePath(ctx, l, fatimaDB, res.SchoolIds)
		if err != nil {
			return 0, 0, fmt.Errorf("updateResourcePath: %s", err)
		}
		totalUpdatedStudentPackage += updatedRowsStudentPackage
		totalUpdatedPackage += updatedRowsPackage
	}
	return totalUpdatedStudentPackage, totalUpdatedPackage, nil
}

func getStudentIDsInStudentPackage(ctx context.Context, l *zap.SugaredLogger, db database.QueryExecer, limit int, latestID string) ([]string, error) {
	studentRows, err := db.Query(ctx, studentIDsScanQuery, latestID, limit)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	if err := studentRows.Err(); err != nil {
		return nil, fmt.Errorf("studentRows.Err(): %s", err)
	}
	defer studentRows.Close()
	var studentIDs []string
	for studentRows.Next() {
		var studentID string
		err = studentRows.Scan(&studentID)
		if err != nil {
			l.Errorf("failed to scan studentID in student package: %s", err.Error())
		}
		studentIDs = append(studentIDs, studentID)
	}
	return studentIDs, nil
}

func updateResourcePath(
	ctx context.Context,
	l *zap.SugaredLogger,
	db database.QueryExecer,
	schoolStudentIDs []*bob_pbv1.SchoolIDWithStudentIDs,
) (int64, int64, error) {
	var totalUpdatedStudentPackage, totalUpdatedPackage int64
	for _, schoolStudentID := range schoolStudentIDs {
		tag, err := db.Exec(ctx, resourcePathUpdateQuery, schoolStudentID.SchoolId, schoolStudentID.StudentIds)
		if err != nil {
			l.Errorf("Error when update student_package: %v", err.Error())
		}
		totalUpdatedStudentPackage += tag.RowsAffected()
		updatedRow, err := getAndUpdatePackageIDFromStudentID(ctx, l, db, schoolStudentID)
		if err != nil {
			return 0, 0, fmt.Errorf("getAndUpdatePackageIDFromStudentID: %s", err)
		}
		totalUpdatedPackage += updatedRow
	}
	return totalUpdatedStudentPackage, totalUpdatedPackage, nil
}

func getAndUpdatePackageIDFromStudentID(ctx context.Context, l *zap.SugaredLogger, db database.QueryExecer, schoolStudentID *bob_pbv1.SchoolIDWithStudentIDs) (int64, error) {
	packageIDRows, err := db.Query(ctx, getPackageIDByStudentQuery, schoolStudentID.StudentIds)
	if err != nil {
		return 0, fmt.Errorf("db.Query: %s", err)
	}
	defer packageIDRows.Close()
	var packageIDs []string
	for packageIDRows.Next() {
		var packageID string
		err = packageIDRows.Scan(&packageID)

		if err != nil {
			l.Errorf("failed to scan packageID in student package: %s", err.Error())
		}
		packageIDs = append(packageIDs, packageID)
	}
	tag, err := db.Exec(ctx, resourcePathUpdateForPackageQuery, schoolStudentID.SchoolId, packageIDs)
	if err != nil {
		l.Errorf("failed to update resourcePath in package: %s", err.Error())
	}
	return tag.RowsAffected(), nil
}

func runMigrateResourcePath(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger().Sugar()
	bobConn := rsc.GRPCDial("bob")
	fatimaDB := rsc.DB()
	updatedStudentPackage, updatedPackage, err := migrateResourcePath(ctx, zapLogger, fatimaDB, bobConn)
	if err != nil {
		return err
	}
	zapLogger.Infof("Migrate result: %d row student package is migrated", updatedStudentPackage)
	zapLogger.Infof("Migrate result: %d row package is migrated", updatedPackage)
	return nil
}

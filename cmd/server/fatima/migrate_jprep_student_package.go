package fatima

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	fatimaCfg "github.com/manabie-com/backend/internal/fatima/configurations"
	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	scanJprepStudentPackagesQuery = `
		SELECT student_packages.student_package_id, 
			student_packages.student_id, 
			student_packages.package_id, 
			student_packages.properties,
			student_packages.start_at, 
			student_packages.end_at,
			student_packages.is_active,
			student_packages.location_ids 
		FROM student_packages
		WHERE student_packages.student_package_id = ANY($1) 
		AND student_packages.location_ids IS NOT NULL
	`

	scanJprepStudentPackagesGroupByStudentIDQuery = `
		SELECT sp.student_id, array_agg(sp.student_package_id)
		FROM student_packages sp
		WHERE sp.deleted_at IS NULL
			AND sp.resource_path = $1
			AND sp.location_ids IS NOT NULL
			AND sp.is_active = TRUE 
			AND end_at >= now()
		GROUP BY sp.student_id
		LIMIT $2 
		OFFSET $3
	`
)

func init() {
	bootstrap.RegisterJob("fatima_migrate_jprep_student_package", RunMigrateJprepStudentPackage)
}

func RunMigrateJprepStudentPackage(ctx context.Context, fatimaCfg fatimaCfg.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	zLogger := rsc.Logger()
	jsm := rsc.NATS()

	return MigrateJprepStudentPackage(ctx, fatimaCfg, db.DB.(*pgxpool.Pool), jsm, zLogger)
}

func MigrateJprepStudentPackage(ctx context.Context, fatimaCfg fatimaCfg.Config, dbPool *pgxpool.Pool, jsm nats.JetStreamManagement, zLogger *zap.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if zLogger == nil {
		zLogger = logger.NewZapLogger(fatimaCfg.Common.Log.ApplicationLevel, fatimaCfg.Common.Environment == "local")
	}

	jprepOrgID := fmt.Sprint(constants.JPREPSchool)

	ctx = auth.InjectFakeJwtToken(ctx, jprepOrgID)
	totalStudents, totalStudentPackages := migrateJprepStudentPackage(ctx, dbPool, jsm, jprepOrgID, zLogger)
	zLogger.Sugar().Infof("There is/are %d new student package migrated from org JPREP (for %d students). ", totalStudentPackages, totalStudents)
	return nil
}

func migrateJprepStudentPackage(ctx context.Context, fatimaDB *pgxpool.Pool, jsm nats.JetStreamManagement, organizationID string, zLogger *zap.Logger) (int, int) {
	if organizationID == "" {
		zLogger.Fatal("running is requires a school id")
		return 0, 0
	}
	// setup job
	const perBatch = 100
	offset := 0
	// scan for student have student_package
	var totalStudents int
	var totalStudentPackages int
	for {
		rows, err := fatimaDB.Query(ctx, scanJprepStudentPackagesGroupByStudentIDQuery, organizationID, perBatch, offset)
		if err != nil {
			zLogger.Sugar().Fatalf("error at querying student_packages (group by student_id): %v", err)
		}
		defer rows.Close()
		offset += perBatch

		mapStudentIDAndStudentPackageIDs := make(map[string][]string)
		for rows.Next() {
			var studentID pgtype.Text
			var studentPackageIDs pgtype.TextArray
			err := rows.Scan(&studentID, &studentPackageIDs)
			if err != nil {
				zLogger.Sugar().Fatalf("error at getStudentPackage: %v", err)
			}

			studentPackageIDsArr := database.FromTextArray(studentPackageIDs)
			mapStudentIDAndStudentPackageIDs[studentID.String] = studentPackageIDsArr
			totalStudentPackages += len(studentPackageIDsArr)
			totalStudents++
		}
		if len(mapStudentIDAndStudentPackageIDs) == 0 {
			zLogger.Sugar().Infof("Query return 0 rows, done migrating")
			break
		}

		mapStudentIDAndStudentPackages, err := getJprepStudentPackage(ctx, fatimaDB, mapStudentIDAndStudentPackageIDs)
		if err != nil {
			zLogger.Sugar().Fatalf("error at getJprepStudentPackage: %v", err)
		}

		err = handlePublishJprepStudentPackageEventForStudent(ctx, jsm, zLogger, mapStudentIDAndStudentPackages)
		if err != nil {
			zLogger.Sugar().Fatalf("error at handlePublishJprepStudentPackageEventForStudent: %v", err)
		}
	}
	return totalStudents, totalStudentPackages
}

func getJprepStudentPackage(ctx context.Context, fatimaDB *pgxpool.Pool, mapStudentIDAndStudentPackageIDs map[string][]string) (map[string][]*fatima_entities.StudentPackage, error) {
	mapStudentIDAndStudentPackages := make(map[string][]*fatima_entities.StudentPackage)
	for studentID, studentPackageIDs := range mapStudentIDAndStudentPackageIDs {
		mapStudentIDAndStudentPackages[studentID] = make([]*fatima_entities.StudentPackage, 0)
		rows, err := fatimaDB.Query(ctx, scanJprepStudentPackagesQuery, studentPackageIDs)
		if err != nil {
			zLogger.Sugar().Fatalf("error at querying student_packages by student_package_id: %v", err)
		}
		for rows.Next() {
			var studentPackageID, studentIDText, packageID pgtype.Text
			var properties pgtype.JSONB
			var isActive pgtype.Bool
			var startAt, endAt pgtype.Timestamptz
			var locationIDs pgtype.TextArray
			err := rows.Scan(&studentPackageID, &studentIDText, &packageID, &properties, &startAt, &endAt, &isActive, &locationIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to scan a row of student_package: %v", err)
			}
			studentPkg := &fatima_entities.StudentPackage{
				ID:          studentPackageID,
				StudentID:   studentIDText,
				PackageID:   packageID,
				Properties:  properties,
				StartAt:     startAt,
				EndAt:       endAt,
				IsActive:    isActive,
				LocationIDs: locationIDs,
			}

			mapStudentIDAndStudentPackages[studentID] = append(mapStudentIDAndStudentPackages[studentID], studentPkg)
		}
	}

	return mapStudentIDAndStudentPackages, nil
}

func handlePublishJprepStudentPackageEventForStudent(ctx context.Context, jsm nats.JetStreamManagement, zLogger *zap.Logger, mapStudentIDAndStudentPackages map[string][]*fatima_entities.StudentPackage) error {
	event := &npb.EventSyncStudentPackage{
		StudentPackages: make([]*npb.EventSyncStudentPackage_StudentPackage, 0),
	}
	for studentID, studentPackages := range mapStudentIDAndStudentPackages {
		studentPkg := &npb.EventSyncStudentPackage_StudentPackage{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:  studentID,
			Packages:   make([]*npb.EventSyncStudentPackage_Package, 0),
		}
		for _, studentPackage := range studentPackages {
			courseIDs, err := studentPackage.GetCourseIDs()
			if err != nil {
				return fmt.Errorf("failed to get courseIDs of student_package %s: %v", studentPackage.ID.String, err)
			}
			studentPkg.Packages = append(studentPkg.Packages, &npb.EventSyncStudentPackage_Package{
				CourseIds: courseIDs,
				StartDate: &timestamppb.Timestamp{
					Seconds: studentPackage.StartAt.Time.Unix(),
				},
				EndDate: &timestamppb.Timestamp{
					Seconds: studentPackage.EndAt.Time.Unix(),
				},
			})
		}
		event.StudentPackages = append(event.StudentPackages, studentPkg)
	}

	publishJprepStudentPackage(ctx, jsm, event, zLogger)
	return nil
}

func publishJprepStudentPackage(ctx context.Context, jsm nats.JetStreamManagement, evt protoreflect.ProtoMessage, zLogger *zap.Logger) {
	msg, _ := proto.Marshal(evt)
	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, constants.SubjectSyncJprepStudentPackageEventNats, msg)
		if err == nil {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(1 * time.Second)
			return true, fmt.Errorf("temporary error jsm.PublishContext: %s", err.Error())
		}
		return false, err
	})
	if err != nil {
		zLogger.Error("jsm.PublishContext failed", zap.Error(err))
	}
}

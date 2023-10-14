package fatima

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	bobCfg "github.com/manabie-com/backend/internal/bob/configurations"
	fatimaCfg "github.com/manabie-com/backend/internal/fatima/configurations"
	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var zLogger *zap.Logger

func initNatsJS(fatimaCfg *fatimaCfg.Config) nats.JetStreamManagement {
	jsm, err := nats.NewJetStreamManagement(fatimaCfg.NatsJS.Address, fatimaCfg.NatsJS.User, fatimaCfg.NatsJS.Password, fatimaCfg.NatsJS.MaxReconnects, fatimaCfg.NatsJS.ReconnectWait, fatimaCfg.NatsJS.IsLocal, zLogger)
	if err != nil {
		zLogger.Sugar().Fatalf("failed to create fatima jetstream management: %s", err)
	}
	jsm.ConnectToJS()
	return jsm
}

func RunMigrateStudentSubscriptions(ctx context.Context, bobCfg *bobCfg.Config, fatimaCfg *fatimaCfg.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", bobCfg.Common.Environment == "local")

	// init natsJS
	jsm := initNatsJS(fatimaCfg)
	defer jsm.Close()

	// init dbs
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

	orgQuery := "select organization_id, name from organizations"
	organizations, err := bobDB.Query(ctx, orgQuery)
	if err != nil {
		zLogger.Fatal("Get orgs failed")
	}
	defer organizations.Close()
	// Migrate with RP
	for organizations.Next() {
		var organizationID, name string
		err := organizations.Scan(&organizationID, &name)
		if err != nil {
			zLogger.Sugar().Errorf("failed to scan an orgs row: %s", err)
		}
		ctx = auth.InjectFakeJwtToken(ctx, organizationID)
		totalCourseStudents := MigrateStudentSubscriptions(ctx, fatimaCfg, fatimaDB, jsm, organizationID)
		zLogger.Sugar().Infof("There is/are %d new student subscriptions migrated from org %s. ", totalCourseStudents, name)
	}
	// Migrate without RP
	ctx = auth.InjectFakeJwtToken(ctx, "")
	totalCourseStudents := MigrateStudentSubscriptionsWithoutResourcePath(ctx, fatimaCfg, fatimaDB, jsm)
	zLogger.Sugar().Infof("There is/are %d new student subscriptions migrated from org NULL. ", totalCourseStudents)
}

func MigrateStudentSubscriptions(ctx context.Context, fatimaCfg *fatimaCfg.Config, fatimaDB *pgxpool.Pool, jsm nats.JetStreamManagement, organizationID string) int {
	if fatimaCfg.Common.Environment != "prod" && fatimaCfg.Common.Environment != "uat" && organizationID == "" {
		zLogger.Fatal("running in non (production/uat) requires a school id")
		return 0
	}
	// setup job
	const perBatch = 100
	offset := 0
	// scan for student_packages
	var totalStudentPackages int
	for {
		rows, err := fatimaDB.Query(ctx, scanStudentPackagesQuery, organizationID, perBatch, offset)
		if err != nil {
			zLogger.Sugar().Fatalf("Error at querying student_packages:%w", err.Error())
		}
		offset += perBatch
		studentPackages := make(map[string]*fatima_entities.StudentPackage)
		for rows.Next() {
			studentPackageEvent, err := CreateStudentPackageEvent(ctx, rows, studentPackages)
			if err != nil {
				zLogger.Sugar().Fatalf("Error at creating studentPackageEvent:%w", err.Error())
			}
			publishStudentPackages(ctx, jsm, studentPackageEvent)
			totalStudentPackages++
		}
		if len(studentPackages) == 0 {
			zLogger.Sugar().Infof("Query return 0 rows, done migrating")
			break
		}
	}
	return totalStudentPackages
}

func MigrateStudentSubscriptionsWithoutResourcePath(ctx context.Context, fatimaCfg *fatimaCfg.Config, fatimaDB *pgxpool.Pool, jsm nats.JetStreamManagement) int {
	// scan for student_packages
	var totalStudentPackages int
	studentPackages := make(map[string]*fatima_entities.StudentPackage)
	rows, err := fatimaDB.Query(ctx, scanStudentPackagesWithNullRPQuery)
	if err != nil {
		zLogger.Sugar().Fatalf("Error at querying student_packages:%w", err.Error())
	}
	for rows.Next() {
		studentPackageEvent, err := CreateStudentPackageEvent(ctx, rows, studentPackages)
		if err != nil {
			zLogger.Sugar().Fatalf("Error at creating studentPackageEvent:%w", err.Error())
		}
		publishStudentPackages(ctx, jsm, studentPackageEvent)
		totalStudentPackages++
	}
	return totalStudentPackages
}

func CreateStudentPackageEvent(ctx context.Context, rows pgx.Rows, studentPackages map[string]*fatima_entities.StudentPackage) (*npb.EventStudentPackage, error) {
	var studentPackageID, studentID, packageID pgtype.Text
	var properties pgtype.JSONB
	var isActive pgtype.Bool
	var startAt, endAt pgtype.Timestamptz
	var locationIDs pgtype.TextArray
	err := rows.Scan(&studentPackageID, &studentID, &packageID, &properties, &startAt, &endAt, &isActive, &locationIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to scan a row of student_package: %s", err)
	}
	if _, ok := studentPackages[studentPackageID.String]; !ok {
		studentPackages[studentPackageID.String] = &fatima_entities.StudentPackage{
			ID:         studentPackageID,
			StudentID:  studentID,
			PackageID:  packageID,
			Properties: properties,
			StartAt:    startAt,
			EndAt:      endAt,
			IsActive:   isActive,
		}
	}
	// get courseIDs from student_packages.properties
	props, err := studentPackages[studentPackageID.String].GetProperties()
	if err != nil {
		return nil, fmt.Errorf("failed to get courseIDs of student_packages: %s", err)
	}
	courseIds := append(props.CanDoQuiz, props.CanViewStudyGuide...)
	courseIds = append(courseIds, props.CanWatchVideo...)

	studentPackageEvent := &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        golibs.Uniq(courseIds),
				StartDate:        timestamppb.New(startAt.Time),
				EndDate:          timestamppb.New(endAt.Time),
				StudentPackageId: studentPackageID.String,
				LocationIds:      database.FromTextArray(locationIDs),
			},
			IsActive: isActive.Bool,
		},
	}
	return studentPackageEvent, nil
}

var (
	scanStudentPackagesQuery = `
SELECT student_packages.student_package_id, 
	   student_packages.student_id, 
	   student_packages.package_id, 
	   student_packages.properties,
	   student_packages.start_at, 
	   student_packages.end_at,
	   student_packages.is_active,
	   student_packages.location_ids 
FROM student_packages
WHERE student_packages.resource_path = $1 
AND student_packages.deleted_at IS NULL 
ORDER BY student_packages.created_at ASC 
LIMIT $2 
OFFSET $3`

	scanStudentPackagesWithNullRPQuery = `
SELECT student_packages.student_package_id, 
	   student_packages.student_id, 
	   student_packages.package_id, 
	   student_packages.properties,
	   student_packages.start_at, 
	   student_packages.end_at,
	   student_packages.is_active,
	   student_packages.location_ids  
FROM student_packages 
WHERE student_packages.resource_path IS NULL
OR student_packages.resource_path = ''
AND student_packages.deleted_at IS NULL 
ORDER BY student_packages.created_at ASC; `
)

func publishStudentPackages(ctx context.Context, jsm nats.JetStreamManagement, evt protoreflect.ProtoMessage) {
	msg, _ := proto.Marshal(evt)
	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, constants.SubjectStudentPackageEventNats, msg)
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

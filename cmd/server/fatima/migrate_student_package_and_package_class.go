package fatima

import (
	"context"
	"fmt"
	"time"

	fatimaCfg "github.com/manabie-com/backend/internal/fatima/configurations"
	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	bootstrap.RegisterJob("fatima_migrate_student_package_and_package_class", RunMigrateStudentPackageAndPackageClass)
}

var (
	scanOrganiationQuery = `
		SELECT organization_id, name 
		FROM organizations
	`

	scanStudentPackagesClassQuery = `
		SELECT class_id, location_id, course_id
		FROM student_package_class 
		WHERE student_package_id = $1
			AND deleted_at IS NULL;
	`

	scanStudentPackagesQueryV2 = `
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
		AND student_packages.location_ids IS NOT NULL
		ORDER BY student_packages.created_at ASC 
		LIMIT $2 
		OFFSET $3
	`
)

func RunMigrateStudentPackageAndPackageClass(ctx context.Context, fatimaCfg *fatimaCfg.Config, rsc *bootstrap.Resources) error {
	l := rsc.Logger().Sugar()
	fatimaDB := rsc.DB()

	organizations, err := fatimaDB.Query(ctx, scanOrganiationQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()

	// Migrate with RP
	for organizations.Next() {
		var organizationID, name string
		err := organizations.Scan(&organizationID, &name)
		if err != nil {
			l.Errorf("failed to scan an orgs row: %w", err)
		}
		ctx = auth.InjectFakeJwtToken(ctx, organizationID)
		totalCourseStudents, err := migrateStudentPackageAndPackageClass(ctx, rsc, organizationID)
		if err != nil {
			return err
		}
		l.Infof("There is/are %d new student course migrated from org %s. ", totalCourseStudents, name)
	}
	return nil
}

func migrateStudentPackageAndPackageClass(ctx context.Context, rsc *bootstrap.Resources, organizationID string) (int, error) {
	if organizationID == "" {
		return 0, fmt.Errorf("missing school id")
	}
	l := rsc.Logger()
	jsm := rsc.NATS()
	fatimaDB := rsc.DB()
	// setup job
	const perBatch = 100
	offset := 0
	// scan for student_packages
	var totalStudentPackages int
	for {
		rows, err := fatimaDB.Query(ctx, scanStudentPackagesQueryV2, organizationID, perBatch, offset)
		if err != nil {
			return 0, fmt.Errorf("error at querying student_packages: %s", err)
		}
		offset += perBatch

		studentPackages := make(map[string]*fatima_entities.StudentPackage)
		for rows.Next() {
			studentPackage, err := getStudentPackage(rows, studentPackages)
			if err != nil {
				return 0, fmt.Errorf("error at getStudentPackage: %s", err)
			}

			mapLocationIDCourseIDAndClassID, err := getMapStudentPackageClass(ctx, fatimaDB, studentPackage.ID.String)
			if err != nil {
				return 0, fmt.Errorf("error at getMapStudentPackageClass: %s", err)
			}

			err = handlePublishUpsertStudentPackageV2Event(ctx, l, jsm, studentPackage, mapLocationIDCourseIDAndClassID)
			if err != nil {
				return 0, fmt.Errorf("error at handlePublishUpsertStudentPackageV2Event: %s", err)
			}

			totalStudentPackages++
		}
		if len(studentPackages) == 0 {
			l.Sugar().Infof("Query return 0 rows, done migrating")
			break
		}
	}
	return totalStudentPackages, nil
}

func getStudentPackage(rows pgx.Rows, studentPackages map[string]*fatima_entities.StudentPackage) (*fatima_entities.StudentPackage, error) {
	var studentPackageID, studentID, packageID pgtype.Text
	var properties pgtype.JSONB
	var isActive pgtype.Bool
	var startAt, endAt pgtype.Timestamptz
	var locationIDs pgtype.TextArray
	err := rows.Scan(&studentPackageID, &studentID, &packageID, &properties, &startAt, &endAt, &isActive, &locationIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to scan a row of student_package: %v", err)
	}
	if _, ok := studentPackages[studentPackageID.String]; !ok {
		studentPackages[studentPackageID.String] = &fatima_entities.StudentPackage{
			ID:          studentPackageID,
			StudentID:   studentID,
			PackageID:   packageID,
			Properties:  properties,
			StartAt:     startAt,
			EndAt:       endAt,
			IsActive:    isActive,
			LocationIDs: locationIDs,
		}
	}

	return studentPackages[studentPackageID.String], nil
}

func getMapStudentPackageClass(ctx context.Context, fatimaDB *database.DBTrace, studentPackageID string) (map[string]string, error) {
	rows, err := fatimaDB.Query(ctx, scanStudentPackagesClassQuery, studentPackageID)
	if err != nil {
		return nil, fmt.Errorf("error at querying student_package_class: %v", err)
	}

	mapLocationIDCourseIDAndClassID := make(map[string]string)
	for rows.Next() {
		var classID, locationID, courseID pgtype.Text
		err := rows.Scan(&classID, &locationID, &courseID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row of student_package_class: %s", err)
		}
		mapLocationIDCourseIDAndClassID[locationID.String+courseID.String] = classID.String
	}

	return mapLocationIDCourseIDAndClassID, nil
}

func handlePublishUpsertStudentPackageV2Event(ctx context.Context, l *zap.Logger, jsm nats.JetStreamManagement, studentPackage *fatima_entities.StudentPackage, mapLocationIDCourseIDAndClassID map[string]string) error {
	locationIDs := database.FromTextArray(studentPackage.LocationIDs)
	courseIDs, err := studentPackage.GetCourseIDs()
	if err != nil {
		return fmt.Errorf("failed to get courseIDs of student_packages: %v", err)
	}

	for _, courseID := range courseIDs {
		for _, locationID := range locationIDs {
			event := &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: studentPackage.StudentID.String,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   courseID,
						StartDate:  timestamppb.New(studentPackage.StartAt.Time),
						EndDate:    timestamppb.New(studentPackage.EndAt.Time),
						ClassId:    "",
						LocationId: locationID,
					},
					IsActive: studentPackage.IsActive.Bool,
				},
			}

			if classID, ok := mapLocationIDCourseIDAndClassID[locationID+courseID]; ok {
				event.StudentPackage.Package.ClassId = classID
			} else {
				event.StudentPackage.Package.ClassId = ""
			}

			publishStudentPackageV2(ctx, l, jsm, event)
		}
	}
	return nil
}

func publishStudentPackageV2(ctx context.Context, l *zap.Logger, jsm nats.JetStreamManagement, evt protoreflect.ProtoMessage) {
	msg, _ := proto.Marshal(evt)
	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, constants.SubjectStudentPackageV2EventNats, msg)
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
		l.Error("jsm.PublishContext failed", zap.Error(err))
	}
}

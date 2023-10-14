package fatima

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/fatima/configurations"
	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

var (
	organizationID string
	bucketName     string
	filePath       string
)

type StudentCourse struct {
	StudentID string
	CourseID  string
}

func init() {
	bootstrap.RegisterJob("fatima_migrate_student_package_and_package_class_by_csv", RunMigrationStudentPackageByCSV).
		StringVar(&organizationID, "organizationID", "", "organization ID to run the job").
		StringVar(&bucketName, "bucketName", "", "bucketName to run the job").
		StringVar(&filePath, "filePath", "", "filePath to run the job")
}

// To run this job, use `fatima migrate-student-package-and-package-class-by-csv -- --organizationID="org-id-value" --bucketName="stag-manabie-backend" --filePath="architecture-upload/student_course.csv"`
func RunMigrationStudentPackageByCSV(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	sugaredLogger := rsc.Logger().Sugar()
	sugaredLogger.Infof("Running on env: %s", c.Common.Environment)
	return MigrationStudentPackageClass(ctx, sugaredLogger, rsc, organizationID, bucketName, filePath)
}

func MigrationStudentPackageClass(
	ctx context.Context,
	sugaredLogger *zap.SugaredLogger,
	rsc *bootstrap.Resources,
	orgID string,
	bucketName string,
	filePath string,
) error {
	err := validateParams(orgID, bucketName, filePath)
	if err != nil {
		return err
	}

	ctx = auth.InjectFakeJwtToken(ctx, orgID)

	studentCoursesCSV, err := readDataFromCsv(ctx, bucketName, filePath)

	if len(studentCoursesCSV) > 0 {
		total, err := migrateStudentPackage(ctx, rsc, orgID, studentCoursesCSV)
		if err != nil {
			return err
		}
		sugaredLogger.Infof("There is/are %d new student course migrated from org %s. ", total, orgID)
	}
	if err != nil {
		return fmt.Errorf("%v orgID: %s, bucketName: %s, filePath: %s", err, orgID, bucketName, filePath)
	}

	return nil
}

func migrateStudentPackage(ctx context.Context, rsc *bootstrap.Resources, organizationID string, studentCourses []*StudentCourse) (int, error) {
	if organizationID == "" {
		return 0, fmt.Errorf("missing school id")
	}
	l := rsc.Logger()
	jsm := rsc.NATS()
	fatimaDB := rsc.DB()
	// setup job
	const perBatch = 100
	var totalStudentPackages int
	batches := batchStudentCourses(studentCourses, perBatch)
	for _, batch := range batches {
		retrieveStudentPackagesQuery := `
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
			AND (student_id, properties->'can_do_quiz'->>0) IN (
		`
		valuesQuery := ""
		for i, row := range batch {
			if i > 0 {
				valuesQuery += ", "
			}
			valuesQuery += fmt.Sprintf("('%s', '%s')", row.StudentID, row.CourseID)
		}
		retrieveStudentPackagesQuery += valuesQuery + ")"
		rows, err := fatimaDB.Query(ctx, retrieveStudentPackagesQuery, organizationID)
		if err != nil {
			return 0, fmt.Errorf("error at querying student_packages: %s", err)
		}

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
	}
	return totalStudentPackages, nil
}

func readDataFromCsv(ctx context.Context, bucketName, filePath string) (studentCourses []*StudentCourse, err error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return studentCourses, fmt.Errorf("failed to create client: %s", err.Error())
	}

	reader, err := client.Bucket(bucketName).Object(filePath).NewReader(ctx)
	if err != nil {
		return studentCourses, fmt.Errorf("failed to open file: %v", err.Error())
	}
	defer reader.Close()
	csvReader := csv.NewReader(reader)
	header, err := csvReader.Read()
	if err != nil {
		return studentCourses, fmt.Errorf("failed to read csv header: %v", err)
	}
	if len(header) != 2 || header[0] != "student_id" || header[1] != "course_id" {
		return studentCourses, fmt.Errorf("invalid csv header: %v", err)
	}
	for {
		record, err := csvReader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return studentCourses, fmt.Errorf("failed to read csv record: %v", err)
		}

		if len(record[0]) > 0 && len(record[1]) > 0 {
			studentCourses = append(studentCourses, &StudentCourse{
				StudentID: record[0],
				CourseID:  record[1],
			})
		}
	}

	return studentCourses, nil
}

func batchStudentCourses(studentCourses []*StudentCourse, batchSize int) [][]*StudentCourse {
	batches := make([][]*StudentCourse, 0)

	for batchSize < len(studentCourses) {
		studentCourses, batches = studentCourses[batchSize:], append(batches, studentCourses[0:batchSize:batchSize])
	}
	batches = append(batches, studentCourses)

	return batches
}

func validateParams(orgID, bucketName, filePath string) error {
	if strings.TrimSpace(orgID) == "" {
		return errors.New("organizationID cannot be empty")
	}

	if strings.TrimSpace(bucketName) == "" {
		return errors.New("bucketName cannot be empty")
	}

	if strings.TrimSpace(filePath) == "" {
		return errors.New("filePath cannot be empty")
	}
	return nil
}

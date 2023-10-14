package communication

import (
	"context"
	"fmt"
	"time"

	cmd_fatima "github.com/manabie-com/backend/cmd/server/fatima"
	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/fatima/configurations"
	fatima_ents "github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/cucumber/godog"
	"go.uber.org/multierr"
)

type JobMigrateStudentPackageAndPackageClassSuite struct {
	*common.NotificationSuite
	mapCourseIDAndClassID          map[string]string
	studentPackagesInserted        []*fatima_ents.StudentPackage
	studentPackagesClassesInserted []*fatima_ents.StudentPackageClass
}

func (c *SuiteConstructor) InitJobMigrateStudentPackageAndPackageClass(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JobMigrateStudentPackageAndPackageClassSuite{
		NotificationSuite:     dep.notiCommonSuite,
		mapCourseIDAndClassID: make(map[string]string),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`:                               s.CreatesNumberOfStudents,
		`^school admin creates "([^"]*)" courses$`:                                s.CreatesNumberOfCourses,
		"^insert class for each course into database$":                            s.insertClassForEachCourseIntoDatabase,
		"^insert student_package and student_package_class into fatima database$": s.insertStudentPackageAndStudentPackageClassIntoFatimaDatabase,
		"^run MigrateStudentPackageAndPackageClass$":                              s.runMigrateStudentCourses,
		"^synced data on bob database correctly$":                                 s.syncedDataOnBobDatabaseCorrectly,
		"^waiting to sync process is finished$":                                   s.waitingToSyncProcessIsFinish,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JobMigrateStudentPackageAndPackageClassSuite) insertClassForEachCourseIntoDatabase(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	classRepo := repo.ClassRepo{}

	classes := make([]*domain.Class, 0)
	for _, course := range commonState.Courses {
		classID := idutil.ULIDNow()
		class := &domain.Class{
			ClassID:      classID,
			Name:         "class-" + classID,
			CourseID:     course.ID,
			LocationID:   commonState.Organization.DefaultLocation.ID,
			SchoolID:     commonState.CurrentResourcePath,
			ResourcePath: commonState.CurrentResourcePath,
		}

		classes = append(classes, class)
		s.mapCourseIDAndClassID[course.ID] = classID
	}

	err := classRepo.UpsertClasses(ctx, s.BobDBConn, classes)
	if err != nil {
		return ctx, fmt.Errorf("classRepo.UpsertClasses: %v", err)
	}

	return ctx, nil
}

func (s *JobMigrateStudentPackageAndPackageClassSuite) insertStudentPackageAndStudentPackageClassIntoFatimaDatabase(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	studentPackageRepo := &repositories.StudentPackageRepo{}
	studentPackageClassRepo := &repositories.StudentPackageClassRepo{}

	startAt := time.Now()
	endAt := time.Now().AddDate(0, 0, 5)
	for _, student := range commonState.Students {
		for _, course := range commonState.Courses {
			sp := &fatima_ents.StudentPackage{}
			spID := idutil.ULIDNow()
			database.AllNullEntity(sp)
			err := multierr.Combine(
				sp.ID.Set(spID),
				sp.StudentID.Set(student.ID),
				sp.PackageID.Set(nil),
				sp.StartAt.Set(startAt),
				sp.LocationIDs.Set([]string{commonState.Organization.DefaultLocation.ID}),
				sp.EndAt.Set(endAt),
				sp.Properties.Set(&fatima_ents.StudentPackageProps{
					CanWatchVideo:     []string{course.ID},
					CanViewStudyGuide: []string{course.ID},
					CanDoQuiz:         []string{course.ID},
					LimitOnlineLesson: 0,
				}),
				sp.IsActive.Set(true),
			)
			if err != nil {
				return ctx, fmt.Errorf("multierr.Combine: %v", err)
			}

			_ = studentPackageRepo.Insert(ctx, s.FatimaDBConn, sp)

			s.studentPackagesInserted = append(s.studentPackagesInserted, sp)

			spc := &fatima_ents.StudentPackageClass{}
			database.AllNullEntity(spc)
			err = multierr.Combine(
				spc.ClassID.Set(s.mapCourseIDAndClassID[course.ID]),
				spc.CourseID.Set(course.ID),
				spc.LocationID.Set(commonState.Organization.DefaultLocation.ID),
				spc.StudentID.Set(student.ID),
				spc.StudentPackageID.Set(spID),
			)
			if err != nil {
				return ctx, fmt.Errorf("multierr.Combine: %v", err)
			}

			s.studentPackagesClassesInserted = append(s.studentPackagesClassesInserted, spc)

			_ = studentPackageRepo.Insert(ctx, s.FatimaDBConn, sp)
			_ = studentPackageClassRepo.BulkUpsert(ctx, s.FatimaDBConn, []*fatima_ents.StudentPackageClass{spc})
		}
	}

	return ctx, nil
}

// setupResourceForRunMigrateStudentCourses setups a resource object so that we can run job functions.
func (s *JobMigrateStudentPackageAndPackageClassSuite) setupResourceForRunMigrateStudentCourses(ctx context.Context, cfg configurations.Config) (*bootstrap.Resources, error) {
	dbName := "fatima"
	rsc := bootstrap.NewResources().WithServiceName(dbName).WithLoggerC(&cfg.Common).WithDatabaseC(ctx, cfg.PostgresV2.Databases).WithNATSC(&cfg.NatsJS)
	return rsc, nil
}

func (s *JobMigrateStudentPackageAndPackageClassSuite) runMigrateStudentCourses(ctx context.Context) (context.Context, error) {
	rsc, err := s.setupResourceForRunMigrateStudentCourses(ctx, fatimaConfig)
	if err != nil {
		return ctx, fmt.Errorf("failed to setup Resources object: %s", err)
	}
	defer rsc.Cleanup() //nolint:errcheck
	err = cmd_fatima.RunMigrateStudentPackageAndPackageClass(ctx, &fatimaConfig, rsc)
	return ctx, err
}

func (s *JobMigrateStudentPackageAndPackageClassSuite) syncedDataOnBobDatabaseCorrectly(ctx context.Context) (context.Context, error) {
	queryStudentPackage := `
		SELECT course_id, student_id, start_at, end_at, created_at, updated_at, deleted_at, location_id
		FROM public.notification_student_courses
		WHERE course_id = $1 
			AND student_id = $2
			AND location_id = $3
			AND deleted_at IS NULL;
	`

	queryStudentPackageClass := `
		SELECT student_id, class_id, start_at, end_at, created_at, updated_at, location_id, course_id, deleted_at
		FROM public.notification_class_members
		WHERE class_id = $1 
			AND student_id = $2;
	`

	err := try.Do(func(attempt int) (bool, error) {
		// check student_package
		for _, sp := range s.studentPackagesInserted {
			if !sp.IsActive.Bool {
				continue
			}

			studentCourse := &entities.NotificationStudentCourse{}
			database.AllNullEntity(studentCourse)

			courseIDs, err := sp.GetCourseIDs()
			if err != nil {
				return false, fmt.Errorf("failed to get courseIDs of student_packages: %v", err)
			}

			for _, courseID := range courseIDs {
				for _, locationID := range database.FromTextArray(sp.LocationIDs) {
					row := s.BobDBConn.QueryRow(ctx, queryStudentPackage, courseID, sp.StudentID, locationID)

					fields := []interface{}{}
					fields = append(fields, database.GetScanFields(studentCourse, database.GetFieldNames(studentCourse))...)
					err := row.Scan(fields...)
					if err != nil {
						if attempt < 10 {
							time.Sleep(2 * time.Second)
							return attempt < 10, err
						}
						return false, fmt.Errorf("failed check student_package: %v", err)
					}

					if studentCourse.CourseID.String != courseID {
						return false, fmt.Errorf("invalid synced student_course with course_id, expected: %s, got: %s", courseID, studentCourse.CourseID.String)
					}

					if studentCourse.StudentID.String != sp.StudentID.String {
						return false, fmt.Errorf("invalid synced student_course with student_id, expected: %s, got: %s", sp.StudentID.String, studentCourse.StudentID.String)
					}

					if studentCourse.LocationID.String != locationID {
						return false, fmt.Errorf("invalid synced student_course with location_id, expected: %s, got: %s", locationID, studentCourse.LocationID.String)
					}

					if studentCourse.StartAt.Time.Truncate(time.Millisecond).String() != sp.StartAt.Time.Truncate(time.Millisecond).String() {
						return false, fmt.Errorf("invalid synced student_course with start_time, expected: %s, got: %s", sp.StartAt.Time.Truncate(time.Millisecond).String(), studentCourse.StartAt.Time.Truncate(time.Millisecond).String())
					}

					if studentCourse.EndAt.Time.Truncate(time.Millisecond).String() != sp.EndAt.Time.Truncate(time.Millisecond).String() {
						return false, fmt.Errorf("invalid synced student_course with end_time, expected: %s, got: %s", sp.EndAt.Time.Truncate(time.Millisecond).String(), studentCourse.EndAt.Time.Truncate(time.Millisecond).String())
					}
				}
			}
		}

		// check student_package_class

		for _, spc := range s.studentPackagesClassesInserted {
			classMember := &entities.NotificationClassMember{}
			database.AllNullEntity(classMember)

			row := s.BobDBConn.QueryRow(ctx, queryStudentPackageClass, spc.ClassID, spc.StudentID)
			fields := []interface{}{}
			fields = append(fields, database.GetScanFields(classMember, database.GetFieldNames(classMember))...)
			err := row.Scan(fields...)
			if err != nil {
				return false, fmt.Errorf("failed check student_package_class: %v", err)
			}

			if classMember.StudentID.String != spc.StudentID.String {
				return false, fmt.Errorf("invalid synced class_member with student_id, expected: %s, got: %s", spc.StudentID.String, classMember.StudentID.String)
			}

			if classMember.ClassID.String != spc.ClassID.String {
				return false, fmt.Errorf("invalid synced class_member with class_id, expected: %s, got: %s", spc.ClassID.String, classMember.ClassID.String)
			}

			if classMember.LocationID.String != spc.LocationID.String {
				return false, fmt.Errorf("invalid synced class_member with location_id, expected: %s, got: %s", spc.LocationID.String, classMember.LocationID.String)
			}
		}
		return false, nil
	})
	if err != nil {
		return ctx, fmt.Errorf("try.Do: %v", err)
	}

	return ctx, nil
}

func (s *JobMigrateStudentPackageAndPackageClassSuite) waitingToSyncProcessIsFinish(ctx context.Context) (context.Context, error) {
	fmt.Printf("\nWaiting for student_package data to be synced...\n")
	waitTime := 20 * time.Second
	time.Sleep(waitTime)
	return ctx, nil
}

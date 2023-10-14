package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/cmd/server/fatima"
	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/features/communication/common/helpers"
	fatima_entities "github.com/manabie-com/backend/internal/fatima/entities"
	fatima_repo "github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
)

type JobMigrateJprepStudenPackageSuite struct {
	*common.NotificationSuite
	studentIDs                         []string
	courseIDs                          []string
	studentPackages                    []*fatima_entities.StudentPackage
	mapExpectedMigratedStudentPackages map[string][]*fatima_entities.StudentPackage // student_id + course_id -> student_package
}

func (c *SuiteConstructor) InitJobMigrateJprepStudentPackage(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JobMigrateJprepStudenPackageSuite{
		NotificationSuite:                  dep.notiCommonSuite,
		mapExpectedMigratedStudentPackages: make(map[string][]*fatima_entities.StudentPackage),
	}

	stepsMapping := map[string]interface{}{
		`^some valid student package in fatima database$`:                s.someValidStudentPackageInFatimaDatabase,
		`^run migration for jprep student package$`:                      s.runMigrationForJPREPStudentPackage,
		`^notification system must store student course data correctly$`: s.notificationSystemMustStoreStudentCourseDataCorrectly,
		`^waiting for notification system sync data$`:                    s.waitingForNotificationSystemSyncData,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JobMigrateJprepStudenPackageSuite) someValidStudentPackageInFatimaDatabase(ctx context.Context) (context.Context, error) {
	numberOfStudent := 5
	numberOfCourse := 5
	for i := 0; i < numberOfStudent; i++ {
		s.studentIDs = append(s.studentIDs, idutil.ULIDNow())
	}
	for i := 0; i < numberOfCourse; i++ {
		s.courseIDs = append(s.courseIDs, idutil.ULIDNow())
	}

	for _, studentID := range s.studentIDs {
		for _, courseID := range s.courseIDs {
			inActiveStudentPackage := &fatima_entities.StudentPackage{
				ID:          database.Text(idutil.ULIDNow()),
				StudentID:   database.Text(studentID),
				LocationIDs: database.TextArray([]string{helpers.JPREPOrgLocation}),
				IsActive:    database.Bool(true),
				StartAt:     database.Timestamptz(time.Now().AddDate(0, -2, 0)),
				EndAt:       database.Timestamptz(time.Now().AddDate(0, -1, 0)),
				Properties: database.JSONB(fatima_entities.StudentPackageProps{
					CanWatchVideo:     []string{courseID},
					CanViewStudyGuide: []string{courseID},
					CanDoQuiz:         []string{courseID},
					LimitOnlineLesson: 0,
				}),
				PackageID: pgtype.Text{Status: pgtype.Null},
			}
			s.studentPackages = append(s.studentPackages, inActiveStudentPackage)

			activeStudentPackage := &fatima_entities.StudentPackage{
				ID:          database.Text(idutil.ULIDNow()),
				StudentID:   database.Text(studentID),
				LocationIDs: database.TextArray([]string{helpers.JPREPOrgLocation}),
				IsActive:    database.Bool(true),
				StartAt:     database.Timestamptz(time.Now().AddDate(0, -1, 0)),
				EndAt:       database.Timestamptz(time.Now().AddDate(0, 1, 0)),
				Properties: database.JSONB(fatima_entities.StudentPackageProps{
					CanWatchVideo:     []string{courseID},
					CanViewStudyGuide: []string{courseID},
					CanDoQuiz:         []string{courseID},
					LimitOnlineLesson: 0,
				}),
				PackageID: pgtype.Text{Status: pgtype.Null},
			}
			s.studentPackages = append(s.studentPackages, activeStudentPackage)
			studentIDAndCourseID := studentID + "," + courseID
			s.mapExpectedMigratedStudentPackages[studentIDAndCourseID] = append(s.mapExpectedMigratedStudentPackages[studentIDAndCourseID], activeStudentPackage)

			activeStudentPackageInFuture := &fatima_entities.StudentPackage{
				ID:          database.Text(idutil.ULIDNow()),
				StudentID:   database.Text(studentID),
				LocationIDs: database.TextArray([]string{helpers.JPREPOrgLocation}),
				IsActive:    database.Bool(true),
				StartAt:     database.Timestamptz(time.Now().AddDate(0, 1, 0)),
				EndAt:       database.Timestamptz(time.Now().AddDate(0, 2, 0)),
				Properties: database.JSONB(fatima_entities.StudentPackageProps{
					CanWatchVideo:     []string{courseID},
					CanViewStudyGuide: []string{courseID},
					CanDoQuiz:         []string{courseID},
					LimitOnlineLesson: 0,
				}),
				PackageID: pgtype.Text{Status: pgtype.Null},
			}
			s.studentPackages = append(s.studentPackages, activeStudentPackageInFuture)
			s.mapExpectedMigratedStudentPackages[studentIDAndCourseID] = append(s.mapExpectedMigratedStudentPackages[studentIDAndCourseID], activeStudentPackageInFuture)
		}
	}

	studentPackageRepo := fatima_repo.StudentPackageRepo{}

	ctxResourcePath := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(int(helpers.JPREPResourcePath)),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
		},
	})

	err := studentPackageRepo.BulkInsert(ctxResourcePath, s.FatimaDBConn, s.studentPackages)
	if err != nil {
		return ctx, fmt.Errorf("can not insert student package: %v", err)
	}

	return ctx, nil
}

func (s *JobMigrateJprepStudenPackageSuite) waitingForNotificationSystemSyncData(ctx context.Context) (context.Context, error) {
	fmt.Printf("\nWaiting for student_package to be synced...\n")
	time.Sleep(30 * time.Second)
	return ctx, nil
}

func (s *JobMigrateJprepStudenPackageSuite) runMigrationForJPREPStudentPackage(ctx context.Context) (context.Context, error) {
	err := fatima.MigrateJprepStudentPackage(ctx, fatimaConfig, s.FatimaDBConn, s.JSM, nil)
	if err != nil {
		return ctx, fmt.Errorf("err run migrate: %v", err)
	}
	return ctx, nil
}

func (s *JobMigrateJprepStudenPackageSuite) notificationSystemMustStoreStudentCourseDataCorrectly(ctx context.Context) (context.Context, error) {
	notificationStudentCourseRepo := repositories.NotificationStudentCourseRepo{}

	ctxResourcePath := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(int(helpers.JPREPResourcePath)),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
		},
	})

	for studentIDAndCourseID, studentPackages := range s.mapExpectedMigratedStudentPackages {
		studentID := strings.Split(studentIDAndCourseID, ",")[0]
		courseID := strings.Split(studentIDAndCourseID, ",")[1]

		filter := repositories.NewFindNotificationStudentCourseFilter()
		_ = filter.StudentID.Set(studentID)
		_ = filter.CourseID.Set(courseID)

		res, err := notificationStudentCourseRepo.Find(ctxResourcePath, s.BobDBConn, filter)

		if err != nil {
			return ctx, fmt.Errorf("notificationStudentCourseRepo.Find: %v", err)
		}

		if len(res) != len(studentPackages) {
			return ctx, fmt.Errorf("migrate failed, expected %d student package migrated, got %d for student: %s and course: %s", len(studentPackages), len(res), studentID, courseID)
		}
	}
	return ctx, nil
}

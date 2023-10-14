package communication

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/entities"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationSyncStudentPackageSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitNotificationSyncStudentPackage(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &NotificationSyncStudentPackageSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`:                                                                  s.CreatesNumberOfStudents,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`:                            s.CreatesNumberOfCoursesWithClass,
		`^assigning course packages to existing students$`:                                                           s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^sync create student courses and class members successfully with "([^"]*)" students and "([^"]*)" courses$`: s.syncStudentCoursesAndClassMembersSuccessfullyWithStudentsAndCourses,
		`^admin edit assigned course packages with start at "([^"]*)" and end at "([^"]*)"$`:                         s.adminEditAssignedCoursePackagesWithStartAtAndEndAt,
		`^sync update student course packages successfully with start at "([^"]*)" and end at "([^"]*)"$`:            s.syncUpdateStudentCoursePackagesSuccessfullyWithStartAtAndEndAt,
		`^waiting for all student course packages to be synced$`:                                                     s.waitingForAllStudentCoursePackageToBeSynced,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *NotificationSyncStudentPackageSuite) genStudentPackageProfileWithStudentCoursePackageID(studentPackageID string, class, location string, startAt, endAt time.Time) *upb.UpsertStudentCoursePackageRequest_StudentPackageProfile {
	return &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
			StudentPackageId: studentPackageID,
		},
		StartTime: timestamppb.New(startAt),
		EndTime:   timestamppb.New(endAt),
		StudentPackageExtra: []*upb.StudentPackageExtra{
			{
				ClassId:    class,
				LocationId: location,
			},
		},
	}
}

func (s *NotificationSyncStudentPackageSuite) syncStudentCoursesAndClassMembersSuccessfullyWithStudentsAndCourses(ctx context.Context, numStudent int64, numCourse int64) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	ctx = util.ContextWithToken(ctx, commonState.CurrentStaff.Token)
	studentIDs := []string{}
	for _, student := range commonState.Students {
		studentIDs = append(studentIDs, student.ID)
	}

	err := try.Do(func(attempt int) (bool, error) {
		var (
			countActualStudentCourse int64
			countActualClassMember   int64
			failCheckStudentCourse   = false
			failCheckClassMember     = false
		)
		queryStudentCourse := `SELECT count(*) FROM notification_student_courses WHERE student_id = ANY($1) AND deleted_at is NULL`

		err := s.BobDBConn.QueryRow(ctx, queryStudentCourse, database.TextArray(studentIDs)).Scan(&countActualStudentCourse)

		if err != nil {
			return false, err
		}

		countExpected := numStudent * numCourse // same countExpected for class member

		if countActualStudentCourse != countExpected {
			failCheckStudentCourse = true
		}

		queryClassMember := `SELECT count(*) FROM notification_class_members WHERE student_id = ANY($1)
		 AND (start_at <= now() AND end_at >= now())`
		err = s.BobDBConn.QueryRow(ctx, queryClassMember, database.TextArray(studentIDs)).Scan(&countActualClassMember)

		if err != nil {
			return false, err
		}

		if countActualClassMember != countExpected {
			failCheckClassMember = true
		}

		if !failCheckStudentCourse && !failCheckClassMember {
			return false, nil // passed
		}

		retry := attempt < 10

		if retry {
			time.Sleep(2 * time.Second)
			return true, fmt.Errorf("sync student course temporarily failed, retrying")
		}

		errMsg := "sync student package failed, lost data: "
		if failCheckStudentCourse {
			errMsg += fmt.Sprintf("student_course(want: %d, got: %d)", countExpected, countActualStudentCourse)
		}

		if failCheckClassMember {
			errMsg += fmt.Sprintf("class_member(want: %d, got: %d)", countExpected, countActualClassMember)
		}

		return false, fmt.Errorf(errMsg)
	})

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *NotificationSyncStudentPackageSuite) adminEditAssignedCoursePackagesWithStartAtAndEndAt(ctx context.Context, startAtString string, endAtString string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	ctx = util.ContextWithToken(ctx, commonState.CurrentStaff.Token)

	startAt, endAt := parseStringStartAtEndAtToTime(startAtString, endAtString)
	for _, student := range commonState.Students {
		studentPackageProfiles := make([]*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile, 0)

		for _, studentPackage := range student.Packages {
			studentPackageProfiles = append(studentPackageProfiles, s.genStudentPackageProfileWithStudentCoursePackageID(studentPackage.ID, studentPackage.ClassID, studentPackage.LocationID, startAt, endAt))
		}

		req := &upb.UpsertStudentCoursePackageRequest{
			StudentId:              student.ID,
			StudentPackageProfiles: studentPackageProfiles,
		}

		_, errResp := upb.NewUserModifierServiceClient(s.UserMgmtGRPCConn).UpsertStudentCoursePackage(ctx, req)
		if errResp != nil {
			return ctx, fmt.Errorf("err adminEditAssignedCoursePackagesWithStartAtAndEndAt: %v", errResp)
		}

		time.Sleep(time.Second)
	}

	return ctx, nil
}

func (s *NotificationSyncStudentPackageSuite) syncUpdateStudentCoursePackagesSuccessfullyWithStartAtAndEndAt(ctx context.Context, startAtString string, endAtString string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, student := range commonState.Students {
		for _, stdPackage := range student.Packages {
			startAt, endAt := parseStringStartAtEndAtToTime(startAtString, endAtString)

			err := s.checkStudentCoursePkgUpdated(ctx, stdPackage.CourseID, student.ID, startAt, endAt)
			if err != nil {
				return ctx, err
			}

			err = s.checkClassMemberPkgUpdated(ctx, stdPackage.ClassID, student.ID, stdPackage.CourseID, startAt, endAt)
			if err != nil {
				return ctx, err
			}
		}
	}
	return ctx, nil
}

func (s *NotificationSyncStudentPackageSuite) checkStudentCoursePkgUpdated(ctx context.Context, courseID, studentID string, startAt, endAt time.Time) error {
	fieldsStudentCourse := database.GetFieldNames(&entities.NotificationStudentCourse{})
	queryStudentCourse := fmt.Sprintf(`SELECT %s FROM notification_student_courses WHERE course_id = $1 AND student_id = $2 AND deleted_at IS NULL;`, strings.Join(fieldsStudentCourse, ","))
	err := try.Do(func(attempt int) (bool, error) {
		var scanResult entities.NotificationStudentCourses

		if err := database.Select(ctx, s.BobDBConn, queryStudentCourse, courseID, studentID).ScanAll(&scanResult); err != nil {
			return false, err
		}

		// check for duplication
		if len(scanResult) > 1 {
			return false, fmt.Errorf("sync has duplicate results of course %s, student %s", courseID, studentID)
		}

		if len(scanResult) == 0 {
			retry := attempt < 10

			if retry {
				time.Sleep(2 * time.Second)

				return true, fmt.Errorf("sync student course temporarily failed, retrying")
			}
			return false, fmt.Errorf("sync or query failed")
		}

		// check time is updated
		startAtDiff := startAt.Sub(scanResult[0].StartAt.Time)
		endAtDiff := endAt.Sub(scanResult[0].EndAt.Time)

		if startAtDiff != 0 || endAtDiff != 0 {
			return false, fmt.Errorf("sync update student course packages failed: startAt & endAt not correctly updated")
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *NotificationSyncStudentPackageSuite) checkClassMemberPkgUpdated(ctx context.Context, classID, studentID, courseID string, startAt, endAt time.Time) error {
	fieldsClassMember := database.GetFieldNames(&entities.NotificationClassMember{})
	queryClassMember := fmt.Sprintf(`SELECT %s FROM notification_class_members WHERE class_id = $1 AND student_id = $2 AND course_id = $3 AND deleted_at IS NULL`, strings.Join(fieldsClassMember, ","))
	err := try.Do(func(attempt int) (bool, error) {
		var scanResult entities.NotificationClassMembers

		if err := database.Select(ctx, s.BobDBConn, queryClassMember, classID, studentID, courseID).ScanAll(&scanResult); err != nil {
			return false, err
		}

		// check for duplication
		if len(scanResult) > 1 {
			return false, fmt.Errorf("sync has duplicate results of class member %s, student %s", classID, studentID)
		}

		if len(scanResult) == 0 {
			retry := attempt < 10
			if retry {
				time.Sleep(2 * time.Second)

				return true, fmt.Errorf("sync class member temporarily failed, retrying")
			}
			return false, fmt.Errorf("sync or query failed")
		}

		// check time is updated
		startAtDiff := startAt.Sub(scanResult[0].StartAt.Time)
		endAtDiff := endAt.Sub(scanResult[0].EndAt.Time)

		if startAtDiff != 0 || endAtDiff != 0 {
			return false, fmt.Errorf("sync update class member packages failed: startAt & endAt not correctly updated")
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *NotificationSyncStudentPackageSuite) waitingForAllStudentCoursePackageToBeSynced(ctx context.Context) (context.Context, error) {
	fmt.Printf("\nWaiting for student course package to be synced...\n")
	time.Sleep(90 * time.Second)
	return ctx, nil
}

func parseStringStartAtEndAtToTime(startAtString, endAtString string) (time.Time, time.Time) {
	timeLayout := "2006-01-02T15:04:05.000Z"
	startAt, err := time.Parse(timeLayout, startAtString)
	if err != nil {
		fmt.Printf("Error parsing startAT:%s", err)
	}
	endAt, err := time.Parse(timeLayout, endAtString)
	if err != nil {
		fmt.Printf("Error parsing endAT:%s", err)
	}
	return startAt, endAt
}

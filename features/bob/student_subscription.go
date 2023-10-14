package bob

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aNumberOfExistingStudents(ctx context.Context, numberOfStudents int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	//create & insert student into DB
	for i := 0; i < numberOfStudents; i++ {
		//init vars
		testULID := idutil.ULIDNow()
		studentID := "bdd_test_student_id_" + testULID
		studentName := "bdd_test_student_name_" + testULID
		ctx, err := s.createStudentWithName(ctx, studentID, studentName)
		if err != nil {
			return ctx, fmt.Errorf("error while creating existing students: %w", err)
		}
		stepState.StudentIds = append(stepState.StudentIds, studentID)
	}
	return StepStateToContext(ctx, stepState), nil

}

func (s *suite) aNumberOfExistingCourses(ctx context.Context, numberOfCourses int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	//create & insert course into DB
	for i := 0; i < numberOfCourses; i++ {
		//init vars
		testULID := idutil.ULIDNow()
		course := entities.Course{}
		database.AllNullEntity(&course)
		if err := multierr.Combine(
			course.ID.Set("bdd_test_course_id_"+testULID),
			course.Name.Set("bdd-test-course-name-"+string(fmt.Sprint(i))),
			course.SchoolID.Set(1),
			course.Grade.Set(10),
			course.CreatedAt.Set(now),
			course.UpdatedAt.Set(now),
		); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value : %w", err)
		}
		cmdTag, err := database.Insert(ctx, &course, s.DB.Exec)
		if err != nil {
			return ctx, err
		}
		if cmdTag.RowsAffected() == 0 {
			return ctx, errors.New("cannot insert student for testing")
		}
		stepState.CourseIDs = append(stepState.CourseIDs, course.ID.String)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assignCoursePackagesToExistingStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()
	reqs := make([]*fpb.AddStudentPackageCourseRequest, 0, len(stepState.StudentIds))
	for _, studentID := range stepState.StudentIds {
		req := &fpb.AddStudentPackageCourseRequest{
			StudentId:   studentID,
			CourseIds:   stepState.CourseIDs,
			StartAt:     timestamppb.New(now),
			EndAt:       timestamppb.New(now.Add(30 * 24 * time.Hour)),
			LocationIds: stepState.LocationIDs,
		}
		reqs = append(reqs, req)
	}
	for _, req := range reqs {
		res, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).AddStudentPackageCourse(contextWithToken(s, ctx), req)
		if err != nil {
			return nil, err
		}
		stepState.StudentPackageID = res.StudentPackageId
	}
	return ctx, nil
}

func (s *suite) syncStudentSubscriptionSuccessfully(ctx context.Context) (context.Context, error) {
	//sleep to make sure NATS sync data successfully
	time.Sleep(5 * time.Second)
	stepState := StepStateFromContext(ctx)
	var count int
	query :=
		` SELECT count(*) FROM lesson_student_subscriptions WHERE course_id = ANY($1) AND student_id = ANY($2) AND deleted_at IS NULL;`
	err := try.Do(func(attempt int) (bool, error) {
		err := s.DB.QueryRow(ctx, query, stepState.CourseIDs, stepState.StudentIds).Scan(&count)
		if err == nil && count > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(5 * time.Second)
			return true, fmt.Errorf("error querying count student subscriptions: %w", err)
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

//functions for testing update start_at, end_at
func (s *suite) anExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	//init vars
	testULID := idutil.ULIDNow()
	id := "bdd_test_manual_student_id_" + testULID
	student := entities.Student{}
	database.AllNullEntity(&student)
	now := time.Now()
	if err := multierr.Combine(
		student.ID.Set(id),
		student.GivenName.Set("bdd_test_manual_student_name_"+testULID),
		student.CurrentGrade.Set(10),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
		student.StudentExternalID.Set("bdd-test-student-external-id-"+testULID),
		student.StudentNote.Set("bdd-test-student-note"),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value : %w", err)
	}

	cmdTag, err := database.Insert(ctx, &student, s.DB.Exec)
	if err != nil {
		return ctx, err
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert student for testing")
	}
	stepState.StudentIds = []string{id}
	return StepStateToContext(ctx, stepState), nil

}

func (s *suite) anExistingCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	//init vars
	testULID := idutil.ULIDNow()
	id := "bdd_test_manual_course_id_" + testULID
	course := entities.Course{}

	database.AllNullEntity(&course)
	now := time.Now()
	if err := multierr.Combine(
		course.ID.Set(id),
		course.Name.Set("bdd_test_manual_course_name_"+testULID),
		course.Grade.Set(10),
		course.SchoolID.Set(1),
		course.CreatedAt.Set(now),
		course.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value : %w", err)
	}

	cmdTag, err := database.Insert(ctx, &course, s.DB.Exec)
	if err != nil {
		return ctx, err
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert course for testing")
	}

	stepState.CourseIDs = []string{id}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) editCoursePackageTime(ctx context.Context, startAtString string, endAtString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startAt, endAt := parseStringStartAtEndAtToTime(startAtString, endAtString)

	req := &fpb.EditTimeStudentPackageRequest{
		StudentPackageId: stepState.StudentPackageID,
		StartAt:          timestamppb.New(startAt),
		EndAt:            timestamppb.New(endAt),
		LocationIds:      stepState.LocationIDs,
	}

	_, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).EditTimeStudentPackage(contextWithToken(s, ctx), req)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (s *suite) syncStudentSubscriptionWithNewStartAtAndEndAtSuccessfully(ctx context.Context, startAtString string, endAtString string) (context.Context, error) {
	//sleep to make sure NATS sync data successfully
	time.Sleep(2 * time.Second)
	stepState := StepStateFromContext(ctx)
	startAt, endAt := parseStringStartAtEndAtToTime(startAtString, endAtString)
	var scanResult entities.StudentSubscriptions
	fields := database.GetFieldNames(&entities.StudentSubscription{})
	query := fmt.Sprintf(` SELECT %s FROM lesson_student_subscriptions WHERE course_id = $1 AND student_id = $2 AND deleted_at IS NULL;`, strings.Join(fields, ","))
	if err := database.Select(ctx, s.DB, query, stepState.CourseIDs[0], stepState.StudentIds[0]).ScanAll(&scanResult); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("DB.QueryRow: %v", err)
	}
	//check for duplication
	if len(scanResult) > 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("sync has duplicate results of course %s, student %s", stepState.CourseIDs[0], stepState.StudentIds[0])
	}
	if len(scanResult) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("sync or query failed")
	}
	//check time is updated
	startAtDiff := startAt.Sub(scanResult[0].StartAt.Time)
	endAtDiff := endAt.Sub(scanResult[0].EndAt.Time)

	if startAtDiff != 0 || endAtDiff != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("sync student subscription failed: startAt & endAt not correctly updated")
	}
	return StepStateToContext(ctx, stepState), nil
}

func parseStringStartAtEndAtToTime(startAtString, endAtString string) (time.Time, time.Time) {
	// parse new time string
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

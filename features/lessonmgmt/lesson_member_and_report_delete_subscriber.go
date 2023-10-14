package lessonmgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) assignCoursePackagesWithStateToExistingStudents(ctx context.Context, state string, locationNumber int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqs := make([]*fpb.AddStudentPackageCourseRequest, 0, len(stepState.StudentIds))
	var startAt, endAt time.Time
	switch state {
	case "active":
		startAt = timeutil.Now().Add(-time.Hour)
		endAt = timeutil.Now().Add(time.Hour)
	case "future":
		startAt = timeutil.Now().Add(time.Hour)
		endAt = timeutil.Now().Add(2 * time.Hour)
	case "inactive":
		startAt = timeutil.Now().Add(-2 * time.Hour)
		endAt = timeutil.Now().Add(-1 * time.Hour)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("status is not available")
	}
	for _, studentID := range stepState.StudentIds {
		// This is for sending the request via gRPC
		req := &fpb.AddStudentPackageCourseRequest{
			StudentId:   studentID,
			CourseIds:   stepState.CourseIDs,
			StartAt:     timestamppb.New(startAt),
			EndAt:       timestamppb.New(endAt),
			LocationIds: []string{stepState.LocationIDs[locationNumber]},
		}
		reqs = append(reqs, req)
	}
	for _, req := range reqs {
		res, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).AddStudentPackageCourse(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
		if err != nil {
			return nil, err
		}
		stepState.StudentPackageID = res.StudentPackageId
	}
	return ctx, nil
}

func (s *Suite) anExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	testULID := idutil.ULIDNow()
	id := "bdd_test_manual_student_id_" + testULID
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}

	name := "student_name" + testULID

	student := &entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)

	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", testULID)),
		student.Email.Set(fmt.Sprintf("email+%s@example.com", testULID)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.SchoolID.Set(stepState.CurrentSchoolID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err := (&repositories.StudentRepo{}).Create(ctx, s.BobDB, student); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentIds = []string{id}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) syncStudentSubscriptionSuccessfully(ctx context.Context) (context.Context, error) {
	// sleep to make sure NATS sync data successfully
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var count int
	query :=
		` SELECT count(*) FROM lesson_student_subscriptions WHERE course_id = ANY($1) AND student_id = ANY($2) AND deleted_at IS NULL;`
	err := try.Do(func(attempt int) (bool, error) {
		err := s.BobDB.QueryRow(ctx, query, stepState.CourseIDs, stepState.StudentIds).Scan(&count)
		if err == nil && count > 0 {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(10 * time.Second)
			return true, fmt.Errorf("error querying count student subscriptions: %w", err)
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreateALessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.CurrentLessonID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing lessonID")
	}
	if len(stepState.FormConfigID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing formConfigID")
	}
	lessonReport := &entities.LessonReport{}
	database.AllNullEntity(lessonReport)
	stepState.LessonReportID = idutil.ULIDNow()
	err := multierr.Combine(
		lessonReport.LessonID.Set(stepState.CurrentLessonID),
		lessonReport.LessonReportID.Set(stepState.LessonReportID),
		lessonReport.FormConfigID.Set(stepState.FormConfigID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson report err: %s", err)
	}
	sql := `INSERT INTO lesson_reports (lesson_id, lesson_report_id, form_config_id, report_submitting_status) VALUES ($1, $2, $3, $4)`
	_, err = s.BobDB.Exec(ctx, sql, lessonReport.LessonID, lessonReport.LessonReportID, lessonReport.FormConfigID, entities.ReportSubmittingStatusSubmitted)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anExistingTypeLesson(ctx context.Context, lessonType string, locationNumber int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lesson := &entities.Lesson{}
	database.AllNullEntity(lesson)
	random := strconv.Itoa(rand.Int())
	lessonID := fmt.Sprintf("lesson_id_%s", random)
	lessonName := fmt.Sprintf("name_%s", random)
	var startAt, endAt time.Time
	classID := idutil.ULIDNow()
	switch lessonType {
	case "past":
		startAt = timeutil.Now().Add(-time.Hour)
		endAt = timeutil.Now().Add(time.Hour)
	case "future":
		startAt = timeutil.Now().Add(time.Hour)
		endAt = timeutil.Now().Add(2 * time.Hour)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("status is not available")
	}
	err := multierr.Combine(
		lesson.LessonID.Set(lessonID),
		lesson.Name.Set(lessonName),
		lesson.CourseID.Set(stepState.CourseIDs[0]),
		lesson.TeacherID.Set(nil),
		lesson.CenterID.Set(stepState.LocationIDs[locationNumber]),
		lesson.CreatedAt.Set(timeutil.Now()),
		lesson.UpdatedAt.Set(timeutil.Now()),
		lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
		lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
		lesson.StreamLearnerCounter.Set(database.Int4(0)),
		lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
		lesson.StartTime.Set(startAt),
		lesson.EndTime.Set(endAt),
		lesson.ClassID.Set(classID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
	}

	if err := lesson.Normalize(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
	}

	cmdTag, err := database.Insert(ctx, lesson, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
	}
	stepState.CurrentLessonID = lesson.LessonID.String
	sql := `INSERT INTO lesson_members (lesson_id, user_id, course_id, created_at, updated_at) VALUES ($1, $2, $3, now(), now())`
	_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(stepState.StudentIds[0]), database.Text(stepState.CourseIDs[0]))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) editCoursePackageLocation(ctx context.Context, state string, location int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var startAt, endAt time.Time
	switch state {
	case "active":
		startAt = timeutil.Now().Add(-time.Hour)
		endAt = timeutil.Now().Add(time.Hour)
	case "future":
		startAt = timeutil.Now().Add(time.Hour)
		endAt = timeutil.Now().Add(2 * time.Hour)
	case "inactive":
		startAt = timeutil.Now().Add(-2 * time.Hour)
		endAt = timeutil.Now().Add(-1 * time.Hour)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("status is not available")
	}
	req := &fpb.EditTimeStudentPackageRequest{
		StudentPackageId: stepState.StudentPackageID,
		StartAt:          timestamppb.New(startAt),
		EndAt:            timestamppb.New(endAt),
		LocationIds:      []string{stepState.LocationIDs[location]},
	}
	studentPackageExtras := []*fpb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
		{
			CourseId:   stepState.CourseIDs[0],
			LocationId: stepState.LocationIDs[location],
			ClassId:    idutil.ULIDNow(),
		},
	}
	req.StudentPackageExtra = studentPackageExtras
	_, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).EditTimeStudentPackage(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return ctx, nil
}

func (s *Suite) lessonMemberDeletedAtState(ctx context.Context, state string) (context.Context, error) {
	// sleep to make sure NATS sync data successfully
	time.Sleep(2 * time.Second)
	stepState := StepStateFromContext(ctx)
	var count int
	condDeleted := "deleted_at is null"
	if state == "deleted" {
		condDeleted = "deleted_at is not null"
	}
	query := fmt.Sprintf("SELECT count(*) FROM lesson_members WHERE course_id = $1 AND user_id = $2 AND lesson_id = $3 AND %s;", condDeleted)

	if err := s.BobDB.QueryRow(ctx, query, stepState.CourseIDs[0], stepState.StudentIds[0], stepState.CurrentLessonID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect lesson_members is %s", state)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonReportDeletedAtState(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	condDeleted := "deleted_at is null"
	if state == "deleted" {
		condDeleted = "deleted_at is not null"
	}
	query := fmt.Sprintf("SELECT count(*) FROM lesson_reports WHERE lesson_report_id = $1 AND %s;", condDeleted)
	if err := s.BobDB.QueryRow(ctx, query, stepState.LessonReportID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect lesson_report is %s", state)
	}
	return StepStateToContext(ctx, stepState), nil
}

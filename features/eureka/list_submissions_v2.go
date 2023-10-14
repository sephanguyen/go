package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	consta "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) listSubmissionsUsingVWithValidLocations(ctx context.Context, arg1 string, arg2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherID := idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).ListSubmissionsV2(contextWithToken(s, ctx), &pb.ListSubmissionsV2Request{
		CourseId:    wrapperspb.String(stepState.CourseID),
		LocationIds: stepState.LocationIDs,
		Start:       timestamppb.New(time.Now().Add(-3 * time.Hour)),
		End:         timestamppb.New(time.Now().Add(3 * time.Hour)),
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listSubmissionsUsingVWithInvalidLocations(ctx context.Context, arg1 string, arg2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherID := idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).ListSubmissionsV2(contextWithToken(s, ctx), &pb.ListSubmissionsV2Request{
		CourseId: wrapperspb.String(stepState.CourseID),
		LocationIds: []string{
			idutil.ULIDNow(),
			idutil.ULIDNow(),
		},
		Start: timestamppb.New(time.Now().Add(-3 * time.Hour)),
		End:   timestamppb.New(time.Now().Add(3 * time.Hour)),
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsListSubmissionsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ListSubmissionsV2Response)
	getKey := func(strs ...string) string {
		return strings.Join(strs, "|")
	}
	m := make(map[string]bool)

	for _, item := range resp.Items {
		m[getKey(item.StudentId, item.AssignmentId)] = true
	}
	for _, studentID := range stepState.StudentIDs {
		for _, assignment := range stepState.Assignments {
			if ok := m[getKey(studentID, assignment.AssignmentId)]; !ok {
				// omit expired student in course id
				if stepState.StudentIDExpired == studentID {
					continue
				}
				return StepStateToContext(ctx, stepState), fmt.Errorf("missing submission, student %v assignment %v", studentID, assignment.AssignmentId)
			}
		}
	}
	// omit expired student in course id
	totalStudent := len(stepState.StudentIDs)
	if stepState.StudentIDExpired != "" {
		totalStudent--
	}
	if len(resp.Items) != totalStudent*len(stepState.Assignments) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("amount of items is wrong, expect %d but got %d", totalStudent*len(stepState.Assignments), len(resp.Items))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentsAddedToCourseInSomeValidLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}
	var (
		courseStudents []*entities.CourseStudent
	)

	courseID := idutil.ULIDNow()
	stepState.CourseID = courseID
	// insert multi user to bob db
	if ctx, err := s.insertMultiUserIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.aValidCourseWithIds(ctx, stepState.StudentIDs, courseID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidCourseStudentBackground: %w", err)
	}
	for _, courseStudent := range stepState.CourseStudents {
		for _, locationID := range locationIDs {
			now := time.Now()
			e := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(e)
			if err := multierr.Combine(
				e.CourseStudentID.Set(courseStudent.ID.String),
				e.CourseID.Set(courseStudent.CourseID.String),
				e.StudentID.Set(courseStudent.StudentID.String),
				e.LocationID.Set(locationID),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err set course student access path: %w", err)
			}
			if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert course student access path: %w", err)
			}
		}
	}
	stepState.CourseStudents = courseStudents
	stepState.LocationIDs = locationIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsAreAssignedAssignmentsInStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.userCreateNewAssignments(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create new assignments:%w", err)
	}
	ctx, err = s.addBookToCourse(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}

	resp, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), &pb.UpsertStudyPlanRequest{
		SchoolId: constants.ManabieSchool,
		Name:     idutil.ULIDNow(),
		CourseId: stepState.CourseID,
		BookId:   stepState.BookID,
		Status:   pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		spie entities.StudyPlanItem
		spe  entities.StudyPlan
	)
	stmt := fmt.Sprintf(`
	WITH TMP AS (
		SELECT study_plan_id
		FROM %s
		WHERE study_plan_id = $1 
		OR master_study_plan_id = $1
	)
	UPDATE %s 
	SET available_from = $2, available_to = $3, start_date = $4	
	WHERE study_plan_id IN(SELECT * FROM TMP)
	`, spe.TableName(), spie.TableName())

	if _, err := s.DB.Exec(ctx,
		stmt,
		&resp.StudyPlanId,
		database.Timestamptz(time.Now().Add(-3*time.Hour)),
		database.Timestamptz(time.Now().Add(3*time.Hour)),
		database.Timestamptz(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update available date: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsSubmitTheirAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		if _, err := s.aValidUser(ctx, studentID, consta.RoleStudent); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
		}
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
		stepState.CurrentStudentID = studentID

		if ctx, err := s.ensureStudentIsCreated(ctx, studentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).
			ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
				CourseId: stepState.CourseID,
				BookId:   stepState.BookID,
			})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, err = s.submitAssignment(ctx, "", "", resp.Contents)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listSubmissionsUsingVWithSomeValidLocationsAndSomeInvalidLocations(ctx context.Context, arg1 string, arg2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherID := idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).ListSubmissionsV2(contextWithToken(s, ctx), &pb.ListSubmissionsV2Request{
		CourseId:    wrapperspb.String(stepState.CourseID),
		LocationIds: append(stepState.LocationIDs, idutil.ULIDNow(), idutil.ULIDNow()),
		Start:       timestamppb.New(time.Now().Add(-3 * time.Hour)),
		End:         timestamppb.New(time.Now().Add(3 * time.Hour)),
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnsListSubmissionsIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ListSubmissionsV2Response)
	if len(resp.Items) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total of item is wrong, expect 0 but got %v", len(resp.Items))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listSubmissionsUsingVWithValidLocationsAndCourseIsNull(ctx context.Context, arg1 string, arg2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherID := idutil.ULIDNow()
	if _, err := s.aValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).ListSubmissionsV2(contextWithToken(s, ctx), &pb.ListSubmissionsV2Request{
		LocationIds: stepState.LocationIDs,
		Start:       timestamppb.New(time.Now().Add(-3 * time.Hour)),
		End:         timestamppb.New(time.Now().Add(3 * time.Hour)),
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListSubmissionsOfStudentsWithRandomLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := idutil.ULIDNow()
	if ctx, err := s.aValidCourseWithIds(ctx, stepState.StudentIDs, courseID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidCourseWithIds: %w", err)
	}
	for _, courseStudent := range stepState.CourseStudents {
		n := rand.Int31n(5) + 1
		for i := int32(1); i <= n; i++ {
			now := time.Now()
			e := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(e)
			if err := multierr.Combine(
				e.CourseStudentID.Set(courseStudent.ID.String),
				e.CourseID.Set(courseStudent.CourseID.String),
				e.StudentID.Set(courseStudent.StudentID.String),
				e.LocationID.Set(idutil.ULIDNow()),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err set course student access path: %w", err)
			}
			if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert course student access path: %w", err)
			}
		}
	}

	now := time.Now()
	cbe := &entities.CoursesBooks{}
	database.AllNullEntity(cbe)
	if err := multierr.Combine(
		cbe.BookID.Set(stepState.BookID),
		cbe.CourseID.Set(courseID),
		cbe.CreatedAt.Set(now),
		cbe.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for course book: %w", err)
	}
	if _, err := database.Insert(ctx, cbe, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to crete course book: %w", err)
	}

	resp, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), &pb.UpsertStudyPlanRequest{
		SchoolId: constants.ManabieSchool,
		Name:     idutil.ULIDNow(),
		CourseId: courseID,
		BookId:   stepState.BookID,
		Status:   pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		spie entities.StudyPlanItem
		spe  entities.StudyPlan
	)
	stmt := fmt.Sprintf(`
	WITH TMP AS (
		SELECT study_plan_id
		FROM %s
		WHERE study_plan_id = $1 
		OR master_study_plan_id = $1
	)
	UPDATE %s 
	SET available_from = $2, available_to = $3, start_date = $4	
	WHERE study_plan_id IN(SELECT * FROM TMP)
	`, spe.TableName(), spie.TableName())

	if _, err := s.DB.Exec(ctx,
		stmt,
		&resp.StudyPlanId,
		database.Timestamptz(time.Now().Add(-3*time.Hour)),
		database.Timestamptz(time.Now().Add(3*time.Hour)),
		database.Timestamptz(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update available date: %w", err)
	}

	for _, studentID := range stepState.StudentIDs {
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
		stepState.CurrentStudentID = studentID
		if ctx, err := s.ensureStudentIsCreated(ctx, studentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = token
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).
			ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
				CourseId: courseID,
				BookId:   stepState.BookID,
			})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, err = s.submitAssignment(ctx, "", "", resp.Contents)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentExpiredInCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.StudentIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no student ids")
	}
	stepState.StudentIDExpired = stepState.StudentIDs[0]
	now := timeutil.Now()
	cmd := `UPDATE course_students SET start_at = $1::TIMESTAMPTZ, end_at = $2::TIMESTAMPTZ WHERE student_id = $3 AND course_id = $4`
	_, err := s.DB.Exec(ctx, cmd, database.Timestamptz(now.Add(-time.Hour*5)), database.Timestamptz(now.Add(-time.Hour*4)), stepState.StudentIDExpired, stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable update to course_students %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

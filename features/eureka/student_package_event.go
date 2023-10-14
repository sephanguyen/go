package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) anStudentPackageWith(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	isActive := false
	if status == "active" {
		isActive = true
	}
	stepState.StudentPackageStatus = isActive

	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}
	stepState.LocationIDs = locationIDs

	e := &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: "student_id-" + idutil.ULIDNow(),
			Package: &npb.EventStudentPackage_Package{
				CourseIds:   stepState.CourseIDs,
				StartDate:   timestamppb.New(time.Now().Add(-time.Hour * 24 * 7)),
				EndDate:     timestamppb.New(time.Now().Add(time.Hour * 24 * 7)),
				LocationIds: stepState.LocationIDs,
			},
			IsActive: isActive,
		},
	}
	stepState.Event = e
	if isActive {
		if ctx, err := s.sendActiveStudentPackageNAT(ctx, e); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		time.Sleep(time.Second * 3)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminAddANewStudentPackageWithAn(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 2)
	isActive := true
	e := &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: "student_id-" + idutil.ULIDNow(),
			Package: &npb.EventStudentPackage_Package{
				CourseIds: stepState.CourseIDs,
				StartDate: timestamppb.New(time.Now().Add(-time.Hour * 24 * 7)),
				EndDate:   timestamppb.New(time.Now().Add(time.Hour * 24 * 7)),
			},
			IsActive: isActive,
		},
	}
	stepState.Event = e
	stepState.StudentPackageStatus = isActive
	data, err := proto.Marshal(e)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
	}

	// this behavior simulator when admin add a new student package it will send to nats and the `eureka` subscribe then handle
	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageEventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	time.Sleep(time.Second * 3)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminToggleStudentPackageStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 3)
	e := stepState.Event.(*npb.EventStudentPackage)
	e.StudentPackage.IsActive = !e.StudentPackage.IsActive
	stepState.StudentPackageStatus = e.StudentPackage.IsActive
	stepState.Event = e
	data, err := proto.Marshal(e)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
	}
	// this behavior simulator when admin toggle a student package it will send to nats and the `eureka` subscribe then handle
	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageEventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	time.Sleep(time.Second * 5)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) sendActiveStudentPackageNAT(ctx context.Context, req *npb.EventStudentPackage) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	data, err := proto.Marshal(req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
	}

	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageEventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToHandleCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.ourSystemHaveToRemoveCorrectly(ctx)
	if !stepState.StudentPackageStatus {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.ourSystemHaveToCreateCorrectly(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) ourSystemHaveToCreateCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.ourSystemHaveToCreateCourseStudent(ctx)
	ctx, err2 := s.ourSystemHaveToCreateNewStudyPlan(ctx)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) ourSystemHaveToRemoveCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.ourSystemHaveToDeleteCourseStudent(ctx)
	ctx, err2 := s.ourSystemHaveToDeleteStudyPlan(ctx)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) ourSystemHaveToCreateCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var studentID string
	var courseIDs []string
	if e, ok := stepState.Event.(*npb.EventStudentPackage); ok {
		courseIDs = e.StudentPackage.Package.CourseIds
		studentID = e.GetStudentPackage().StudentId
	} else {
		studentID = stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().StudentId
		courseIDs = []string{stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().Package.CourseId}
	}
	count := 0
	query := `SELECT count(*) FROM course_students WHERE student_id = $1 AND deleted_at IS NULL`
	if err := s.DB.QueryRow(context.Background(), query, studentID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query: %w", err)
	}
	if count != len(courseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create: expected %d, got %d", len(courseIDs), count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToCreateNewStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var studentID string
	var courseIDs []string
	if e, ok := stepState.Event.(*npb.EventStudentPackage); ok {
		courseIDs = e.StudentPackage.Package.CourseIds
		studentID = e.GetStudentPackage().StudentId
	} else {
		studentID = stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().StudentId
		courseIDs = []string{stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().Package.CourseId}
	}
	r := repositories.CourseStudyPlanRepo{}
	courseStudyPlan, err := r.FindByCourseIDs(ctx, s.DB, database.TextArray(courseIDs))
	studyPlanIDs := make([]string, 0, len(courseStudyPlan))

	for _, csp := range courseStudyPlan {
		studyPlanIDs = append(studyPlanIDs, csp.StudyPlanID.String)
	}

	ctx, count, err := s.getNumberOfNewStudyPlanForEachCoursesStudent(ctx, studentID, studyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != len(studyPlanIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect student %s has %d study plan but got %d", studentID, len(studyPlanIDs), count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToDeleteStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var studentID string
	var courseIDs []string
	if e, ok := stepState.Event.(*npb.EventStudentPackage); ok {
		courseIDs = e.StudentPackage.Package.CourseIds
		studentID = e.GetStudentPackage().StudentId
	} else {
		studentID = stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().StudentId
		courseIDs = []string{stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().Package.CourseId}
	}

	ctx, count, err := s.getNumberOfCourseStudentStudyPlan(ctx, studentID, courseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable retrieve number of study plan items")
	}
	if count != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable study plan related to student package")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToDeleteCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var studentID string
	var courseIDs []string
	if e, ok := stepState.Event.(*npb.EventStudentPackage); ok {
		courseIDs = e.StudentPackage.Package.CourseIds
		studentID = e.GetStudentPackage().StudentId
	} else {
		studentID = stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().StudentId
		courseIDs = []string{stepState.Event.(*npb.EventStudentPackageV2).GetStudentPackage().Package.CourseId}
	}
	count := 0
	query := `SELECT count(*) FROM course_students WHERE student_id = $1 AND deleted_at IS NOT NULL`
	if err := s.DB.QueryRow(context.Background(), query, studentID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query: %w", err)
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to delete course_students course = %v, student = %v", courseIDs, studentID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) courseStudentAccessPathsWereCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := try.Do(func(attempt int) (bool, error) {
		count := 0
		query := `SELECT count(*) FROM course_students_access_paths WHERE location_id = ANY($1) AND deleted_at IS NULL`
		if err := s.DB.QueryRow(context.Background(), query, stepState.LocationIDs).Scan(&count); err != nil {
			return true, fmt.Errorf("unable to scan query: %w", err)
		}

		if count == 0 {
			time.Sleep(1 * time.Second)
			return attempt < 5, fmt.Errorf("expected count > 0")
		}
		return false, nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminAddAStudentPackageAndUpdateLocationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(time.Second * 2)
	locationID1 := idutil.ULIDNow()
	locationID2 := idutil.ULIDNow()
	locationIDs := []string{locationID1, locationID2, locationID1}

	stepState.LocationIDs = locationIDs

	// add course_student with location 1
	// update course_student with location 2
	// update course_student with location 1
	studentID := "student_id-" + idutil.ULIDNow()
	stepState.StudentID = studentID
	for i := 0; i < 3; i++ {
		isActive := true
		e := &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: studentID,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:   []string{stepState.CourseIDs[0]},
					StartDate:   timestamppb.New(time.Now().Add(-time.Hour * 24 * 7)),
					EndDate:     timestamppb.New(time.Now().Add(time.Hour * 24 * 7)),
					LocationIds: []string{locationIDs[i]},
				},
				IsActive: isActive,
			},
		}
		stepState.Event = e
		stepState.StudentPackageStatus = isActive
		data, err := proto.Marshal(e)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
		}

		// this behavior simulator when admin add a new student package it will send to nats and the `eureka` subscribe then handle
		_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageEventNats, data)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
		}
		time.Sleep(time.Second * 3)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToUpdatedCourseStudentAccessPathsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := stepState.Event.(*npb.EventStudentPackage)
	var deletedAt pgtype.Timestamptz
	query := `SELECT deleted_at FROM course_students_access_paths WHERE student_id = $1 AND course_id = $2 AND location_id = $3`
	if err := s.DB.QueryRow(context.Background(), query, e.StudentPackage.StudentId, e.StudentPackage.Package.CourseIds[0], stepState.LocationIDs[0]).Scan(&deletedAt); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query: %w", err)
	}

	if deletedAt.Status == pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("deleted_at of location_id %s must be null", stepState.LocationIDs[0])
	}

	if err := s.DB.QueryRow(context.Background(), query, e.StudentPackage.StudentId, e.StudentPackage.Package.CourseIds[0], stepState.LocationIDs[1]).Scan(&deletedAt); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query: %w", err)
	}

	if deletedAt.Status != pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("deleted_at of location_id %s must not be null", stepState.LocationIDs[0])
	}

	return StepStateToContext(ctx, stepState), nil
}

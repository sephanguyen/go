package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aStudentPackageVWith(ctx context.Context, arg1 int, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	isActive := false
	if status == "active" {
		isActive = true
	}
	stepState.StudentPackageStatus = isActive

	id := idutil.ULIDNow()
	stepState.LocationIDs = []string{id}

	e := &npb.EventStudentPackageV2{
		StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
			StudentId: "student_id-" + idutil.ULIDNow(),
			Package: &npb.EventStudentPackageV2_PackageV2{
				CourseId:   stepState.CourseIDs[0],
				StartDate:  timestamppb.New(time.Now().Add(-time.Hour * 24 * 7)),
				EndDate:    timestamppb.New(time.Now().Add(time.Hour * 24 * 7)),
				LocationId: id,
			},
			IsActive: isActive,
		},
	}

	stepState.Event = e
	if isActive {
		if ctx, err := s.sendActiveStudentPackageV2NAT(ctx, e); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		time.Sleep(time.Second * 3)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) sendActiveStudentPackageV2NAT(ctx context.Context, req *npb.EventStudentPackageV2) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	data, err := proto.Marshal(req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
	}

	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageV2EventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminAddANewStudentPackageVWithAPackageOrCourses(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 2)
	isActive := true
	e := &npb.EventStudentPackageV2{
		StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
			StudentId: "student_id-" + idutil.ULIDNow(),
			Package: &npb.EventStudentPackageV2_PackageV2{
				ClassId:   idutil.ULIDNow(),
				CourseId:  stepState.CourseIDs[0],
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
	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageV2EventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	time.Sleep(time.Second * 3)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminAddAStudentPackageVAndUpdateLocation_id(ctx context.Context, arg1 int) (context.Context, error) {
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
		e := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: studentID,
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   stepState.CourseIDs[0],
					StartDate:  timestamppb.New(time.Now().Add(-time.Hour * 24 * 7)),
					EndDate:    timestamppb.New(time.Now().Add(time.Hour * 24 * 7)),
					LocationId: locationIDs[i],
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
		_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageV2EventNats, data)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
		}
		time.Sleep(time.Second * 3)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theAdminToggleStudentPackageVStatus(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 3)
	e := stepState.Event.(*npb.EventStudentPackageV2)
	e.StudentPackage.IsActive = !e.StudentPackage.IsActive
	stepState.StudentPackageStatus = e.StudentPackage.IsActive
	stepState.Event = e
	data, err := proto.Marshal(e)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to marshal data: %w", err)
	}
	// this behavior simulator when admin toggle a student package it will send to nats and the `eureka` subscribe then handle
	_, err = s.JSM.PublishContext(context.Background(), constants.SubjectStudentPackageV2EventNats, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish to nats: %w", err)
	}
	time.Sleep(time.Second * 5)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) courseStudentAccessPathsWereCreatedForV(ctx context.Context, arg1 int) (context.Context, error) {
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

func (s *suite) ourSystemHaveToHandleStudentPackageVCorrectly(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.ourSystemHaveToRemoveStudentPackageVCorrectly(ctx)

	if !stepState.StudentPackageStatus {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.ourSystemHaveToCreateStudentPackageVCorrectly(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) ourSystemHaveToCreateStudentPackageVCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.ourSystemHaveToCreateCourseStudent(ctx)
	ctx, err2 := s.ourSystemHaveToCreateNewStudyPlan(ctx)

	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}
func (s *suite) ourSystemHaveToRemoveStudentPackageVCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.ourSystemHaveToDeleteCourseStudent(ctx)
	ctx, err2 := s.ourSystemHaveToDeleteStudyPlan(ctx)

	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) ourSystemHaveToUpdatedCourseStudentAccessPathsCorrectlyForV(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := stepState.Event.(*npb.EventStudentPackageV2)
	var deletedAt pgtype.Timestamptz
	query := `SELECT deleted_at FROM course_students_access_paths WHERE student_id = $1 AND course_id = $2 AND location_id = $3`
	if err := s.DB.QueryRow(context.Background(), query, e.StudentPackage.StudentId,
		e.StudentPackage.Package.CourseId, stepState.LocationIDs[0]).Scan(&deletedAt); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query : %w", err)
	}

	if deletedAt.Status == pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("deleted_at of location_id %s must be null", stepState.LocationIDs[0])
	}

	if err := s.DB.QueryRow(context.Background(), query, e.StudentPackage.StudentId,
		e.StudentPackage.Package.CourseId, stepState.LocationIDs[1]).Scan(&deletedAt); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to scan query : %w", err)
	}

	if deletedAt.Status != pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("deleted_at of location_id %s must not be null", stepState.LocationIDs[0])
	}

	return StepStateToContext(ctx, stepState), nil
}

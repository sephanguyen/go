package eureka

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userListClassByCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.userListClass(ctx, stepState.CourseID, nil)
}

func (s *suite) userListClassByCourseAndLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.userListClass(ctx, stepState.CourseID, stepState.ClassLocationIDs)
}

func (s *suite) userListClassByCourseAndNotExistLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.userListClass(ctx, stepState.CourseID, []string{"location-not-exist", "location-not-exist-2"})
}

func (s *suite) userListClass(ctx context.Context, courseID string, locations []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.ListClassByCourseRequest{
		CourseId:    courseID,
		LocationIds: locations,
	}

	stepState.Response, stepState.ResponseErr = pb.NewCourseReaderServiceClient(s.Conn).ListClassByCourse(contextWithToken(s, ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedIn(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidToken(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidToken: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustReturnCorrectListOfClassIds(ctx context.Context) (context.Context, error) {
	return s.eurekaReturnClassIds(ctx, "correct")
}

func (s *suite) eurekaReturnEmptyListOfClassIds(ctx context.Context) (context.Context, error) {
	return s.eurekaReturnClassIds(ctx, "empty")
}

func (s *suite) eurekaReturnNilListOfClassIds(ctx context.Context) (context.Context, error) {
	return s.eurekaReturnClassIds(ctx, "nil")
}

func (s *suite) eurekaReturnClassIds(ctx context.Context, expected string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.ListClassByCourseResponse)
	switch expected {
	case "nil":
		if rsp.ClassIds != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected null list classes but got %d class", len(rsp.ClassIds))
		}
	case "empty":
		if len(rsp.ClassIds) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected empty list classes but got %d class", len(rsp.ClassIds))
		}
	case "correct":
		query := "SELECT count(*) FROM course_classes WHERE course_id = $1"
		var count int
		if err := s.DB.QueryRow(ctx, query, &stepState.CourseID).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != len(rsp.ClassIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d class but got %d class", count, len(rsp.ClassIds))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) deleteAllClassesBelongToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := "UPDATE class SET deleted_at = now() WHERE course_id = $1"
	_, err := s.BobDBTrace.Exec(ctx, query, stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

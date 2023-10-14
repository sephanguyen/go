package eurekav2

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/protobuf/proto"
)

func (s *suite) generateCourses(numberOfCourses int, template *epb.UpsertCoursesRequest_Course) []*epb.UpsertCoursesRequest_Course {
	if template == nil {
		// A valid create course req template
		template = &epb.UpsertCoursesRequest_Course{
			Name:   "Course 1",
			BookId: s.BookID,
		}
	}

	courses := make([]*epb.UpsertCoursesRequest_Course, numberOfCourses)
	for i := 0; i < numberOfCourses; i++ {
		courses[i] = proto.Clone(template).(*epb.UpsertCoursesRequest_Course)
	}

	return courses
}

func (s *suite) createAnEmptyCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	reqCourses := s.generateCourses(1, nil)
	resp, err := epb.NewCourseServiceClient(s.Connections.EurekaConn).UpsertCourses(ctx, &epb.UpsertCoursesRequest{
		Courses: reqCourses,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not create course: %w", err)
	}
	stepState.CourseID = resp.CourseIds[0]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewCourses(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfCourses := rand.Intn(20) + 1 //nolint:gosec
	var reqCourses []*epb.UpsertCoursesRequest_Course
	switch validity {
	case "valid":
		reqCourses = s.generateCourses(numberOfCourses, nil)
	case "invalid":
		reqCourses = s.generateCourses(numberOfCourses, &epb.UpsertCoursesRequest_Course{
			Name: "",
		})
	}

	stepState.Response, stepState.ResponseErr = epb.NewCourseServiceClient(s.EurekaConn).UpsertCourses(ctx, &epb.UpsertCoursesRequest{
		Courses: reqCourses,
	})
	if stepState.ResponseErr == nil {
		stepState.CourseIDs = stepState.Response.(*epb.UpsertCoursesResponse).CourseIds
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfCourses := rand.Intn(20) + 1 //nolint:gosec
	reqCourses := s.generateCourses(numberOfCourses, nil)

	res, err := epb.NewCourseServiceClient(s.EurekaConn).UpsertCourses(ctx, &epb.UpsertCoursesRequest{
		Courses: reqCourses,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to seeding some courses: %w", err)
	}
	stepState.CourseIDs = res.CourseIds

	query := "SELECT course_id, updated_at FROM courses WHERE course_id = ANY($1)"
	rows, err := s.EurekaDB.Query(ctx, query, stepState.CourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query course: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		course := new(domain.Course)
		if err := rows.Scan(&course.ID, &course.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to scan course: %w", err)
		}
		stepState.Courses = append(stepState.Courses, *course)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateCourses(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	// Take first <randomly> courses to update, 1 at least
	numberOfUpdates := rand.Intn(len(stepState.CourseIDs)) //nolint:gosec
	if numberOfUpdates == 0 {
		numberOfUpdates = 1
	}

	stepState.UpdatedCourseIDs = append(stepState.UpdatedCourseIDs, stepState.CourseIDs[0:numberOfUpdates]...)
	reqCourses := make([]*epb.UpsertCoursesRequest_Course, 0)
	// Update data by each case.
	switch validity {
	case "valid":
		for _, courseID := range stepState.UpdatedCourseIDs {
			reqCourses = append(reqCourses, &epb.UpsertCoursesRequest_Course{
				CourseId: courseID,
				Name:     "Course 1",
			})
		}
	case "invalid":
		for _, courseID := range stepState.UpdatedCourseIDs {
			reqCourses = append(reqCourses, &epb.UpsertCoursesRequest_Course{
				CourseId: courseID,
				Name:     "",
			})
		}
	}

	stepState.Response, stepState.ResponseErr = epb.NewCourseServiceClient(s.Connections.EurekaConn).UpsertCourses(ctx, &epb.UpsertCoursesRequest{
		Courses: reqCourses,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existedCourseMap := make(map[string]domain.Course)
	for _, course := range stepState.Courses {
		existedCourseMap[course.ID] = course
	}

	query := "SELECT course_id, updated_at FROM courses WHERE course_id = ANY($1)"
	rows, err := s.EurekaDB.Query(ctx, query, stepState.UpdatedCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		course := new(domain.Course)
		if err := rows.Scan(&course.ID, &course.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query course: %w", err)
		}
		existedCourse, ok := existedCourseMap[course.ID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course is missing course_id: %s", course.ID)
		}
		if existedCourse.UpdatedAt.Equal(course.UpdatedAt) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course was not updated course_id: %s", course.ID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfCourses := rand.Intn(20) + 1 //nolint:gosec
	reqCourses := s.generateCourses(numberOfCourses, nil)
	stepState.Response, stepState.ResponseErr = epb.NewCourseServiceClient(s.Connections.EurekaConn).UpsertCourses(ctx, &epb.UpsertCoursesRequest{
		Courses: reqCourses,
	})
	return StepStateToContext(ctx, stepState), nil
}

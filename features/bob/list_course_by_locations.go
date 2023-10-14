package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

func (s *suite) userListCoursesByLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)

	paging := &cpb.Paging{
		Limit: 5,
	}

	stepState.ResponseErr = nil
	stepState.PaginatedCourses = nil
	for {
		stepState.Request.(*bpb.ListCoursesByLocationsRequest).Paging = paging
		resp, err := bpb.NewCourseReaderServiceClient(s.Conn).ListCoursesByLocations(ctx, stepState.Request.(*bpb.ListCoursesByLocationsRequest))
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		if len(resp.Items) == 0 {
			break
		}
		if len(resp.Items) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total course: got: %d, want: %d", len(resp.Items), paging.Limit)
		}
		stepState.PaginatedCourses = append(stepState.PaginatedCourses, resp.Items)
		paging = resp.NextPage
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListCoursesbyLocationRequestMessageSchool(ctx context.Context, arg1, keyword string, location int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationIDs := []string{}
	if location != 0 {
		locationIDs = stepState.CenterIDs[0:location]
	}
	if arg1 == "manabie" {
		stepState.Request = &bpb.ListCoursesByLocationsRequest{Filter: &cpb.CommonFilter{SchoolId: constant.ManabieSchool}, LocationIds: locationIDs, Keyword: keyword}
	} else {
		stepState.Request = &bpb.ListCoursesByLocationsRequest{LocationIds: locationIDs, Keyword: keyword}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listCourseAccessPathExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseAccessPath := repositories.CourseAccessPathRepo{}
	courseAPs := []*entities.CourseAccessPath{}
	for _, course := range stepState.courseIds {
		for _, location := range stepState.CenterIDs {
			item, err := toCourseAccessPathEntity(location, course)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err toCourseAccessPathEntity")
			}
			courseAPs = append(courseAPs, item)
		}
	}
	if err := courseAccessPath.Upsert(ctx, s.DB, courseAPs); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseAccessPath.Upsert")
	}
	return StepStateToContext(ctx, stepState), nil
}

func toCourseAccessPathEntity(locationID, courseID string) (*entities.CourseAccessPath, error) {
	cap := &entities.CourseAccessPath{}
	database.AllNullEntity(cap)
	err := multierr.Combine(
		cap.CourseID.Set(courseID),
		cap.LocationID.Set(locationID),
	)
	if err != nil {
		return nil, err
	}
	return cap, nil
}

func (s *suite) locationsOfCoursesMatchingWithCourseAccessPath(ctx context.Context, location int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if location == 0 && len(stepState.PaginatedCourses) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect 0 courses, got: %d", len(stepState.PaginatedCourses[0]))
	} else {
		courseAccessPath := repositories.CourseAccessPathRepo{}
		for _, courses := range stepState.PaginatedCourses {
			for _, course := range courses {
				if course.Info.Id != "" {
					mapLocationIDsByCourseID, err := courseAccessPath.FindByCourseIDs(ctx, db, []string{course.Info.Id})
					if err != nil {
						return ctx, errors.New("courseAccessPath.FindByCourseIDs")
					}
					found := false
					for _, e := range stepState.CenterIDs[0:location] {
						found = checkLocationsOfCourse(mapLocationIDsByCourseID[course.Info.Id], e)
					}
					if !found {
						return StepStateToContext(ctx, stepState), fmt.Errorf("Location of course %s not matching", course.Info.Id)
					}
				}
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func checkLocationsOfCourse(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

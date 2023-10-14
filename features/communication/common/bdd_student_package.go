package common

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
)

func (s *NotificationSuite) AddPackagesDataOfThoseCoursesForEachStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.Students) == 0 {
		return ctx, fmt.Errorf("not found students to add package")
	}

	if len(stepState.Courses) == 0 {
		return ctx, fmt.Errorf("not found courses to add package")
	}

	if stepState.Organization.DefaultLocation.ID == "" {
		return ctx, fmt.Errorf("not found location to add package")
	}

	for _, student := range stepState.Students {
		err := try.Do(func(attempts int) (bool, error) {
			if err := s.AddCourseToStudentWithLocation(stepState.Organization.Staffs[0], student, stepState.Courses, stepState.MapCourseIDAndStudentIDs, []string{stepState.Organization.DefaultLocation.ID}); err != nil {
				if attempts < 5 {
					fmt.Printf("Got error when AddCourseToStudentWithLocation, trying again after 10s...")
					time.Sleep(10 * time.Second)
					return true, nil
				}
				return false, err
			}

			return false, nil
		})

		if err != nil {
			return ctx, fmt.Errorf("s.AddPackagesDataOfThoseCoursesForEachStudent: %v", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) AddPackagesDataOfThoseCoursesForRandomStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.Students) == 0 {
		return ctx, fmt.Errorf("not found students to add package")
	}

	if len(stepState.Courses) == 0 {
		return ctx, fmt.Errorf("not found courses to add package")
	}

	if stepState.Organization.DefaultLocation.ID == "" {
		return ctx, fmt.Errorf("not found location to add package")
	}

	for i, student := range stepState.Students {
		if i == 0 || rand.Intn(2) == 1 {
			if err := s.AddCourseToStudentWithLocation(stepState.Organization.Staffs[0], student, stepState.Courses, stepState.MapCourseIDAndStudentIDs, []string{stepState.Organization.DefaultLocation.ID}); err != nil {
				return ctx, fmt.Errorf("s.AddPackagesDataOfThoseCoursesForEachStudent: %v", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

func (s *NotificationSuite) CreatesNumberOfCourses(ctx context.Context, num string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numCourse := 0
	if num == "random" {
		numCourse = RandRangeIn(2, 4)
	} else {
		var err error
		if numCourse, err = strconv.Atoi(num); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfCourses: %v", err)
		}
	}

	// Create courses
	courses, err := s.CreateCourses(stepState.Organization.Staffs[0], stepState.Organization.ID, numCourse, []string{stepState.Organization.DefaultLocation.ID})
	if err != nil {
		return ctx, fmt.Errorf("s.CreatesNumberOfCourses: %v", err)
	}

	stepState.Courses = append(stepState.Courses, courses...)

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CreatesNumberOfCoursesWithClass(ctx context.Context, numCourse, numClass string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	intNumCourse := 0
	if numCourse == "random" {
		intNumCourse = RandRangeIn(2, 4)
	} else {
		var err error
		if intNumCourse, err = strconv.Atoi(numCourse); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfCoursesWithClass: %v", err)
		}
	}

	intNumClass := 0
	if numClass == "random" {
		intNumClass = RandRangeIn(2, 4)
	} else {
		var err error
		if intNumClass, err = strconv.Atoi(numClass); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfCoursesWithClass: %v", err)
		}
	}

	// Create courses
	courses, classes, err := s.CreateCoursesWithClass(stepState.Organization.Staffs[0], stepState.Organization.ID, intNumCourse, intNumClass, []string{stepState.Organization.DefaultLocation.ID})
	if err != nil {
		return ctx, fmt.Errorf("s.CreatesNumberOfCourses: %v", err)
	}

	stepState.Courses = append(stepState.Courses, courses...)
	stepState.Classes = append(stepState.Classes, classes...)

	return StepStateToContext(ctx, stepState), nil
}

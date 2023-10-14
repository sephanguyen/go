package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/communication/common/helpers"

	"github.com/pkg/errors"
)

func (s *NotificationSuite) CreatesNumberOfStudents(ctx context.Context, num string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numStudent := 0
	if num == "random" {
		numStudent = RandRangeIn(2, 5)
	} else {
		var err error
		if numStudent, err = strconv.Atoi(num); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudent: %v", err)
		}
	}

	// Create students
	for i := 0; i < numStudent; i++ {
		gradeIdx := RandRangeIn(0, len(stepState.GradeMasters))
		gradeMaster := stepState.GradeMasters[gradeIdx]

		// randomly select a school for student
		idxSchool := RandRangeIn(0, 5)
		school := stepState.Schools[idxSchool]
		stepState.CurrentSchools = append(stepState.CurrentSchools, school)
		newStudent, err := s.CreateStudent(stepState.Organization.Staffs[0], gradeMaster, []string{stepState.Organization.DefaultLocation.ID}, false, 1, school.ID)
		if err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudent: %v", err)
		}
		stepState.Students = append(stepState.Students, newStudent)
		stepState.GradeAssigneds = append(stepState.GradeAssigneds, gradeMaster)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CreatesNumberOfStudentsWithParentsInfo(ctx context.Context, numStudentReq, numParentReq string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numStudent := 0
	if numStudentReq == "random" {
		numStudent = RandRangeIn(2, 5)
	} else {
		var err error
		if numStudent, err = strconv.Atoi(numStudentReq); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
	}

	numParent := 0
	if numParentReq == "random" {
		numParent = RandRangeIn(1, 2)
	} else {
		var err error
		if numParent, err = strconv.Atoi(numParentReq); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
	}

	// Create students
	for studentIdx := 0; studentIdx < numStudent; studentIdx++ {
		gradeIdx := RandRangeIn(0, len(stepState.GradeMasters))
		gradeMaster := stepState.GradeMasters[gradeIdx]

		// randomly select a school for student
		idxSchool := RandRangeIn(1, 5)
		school := stepState.Schools[idxSchool]
		stepState.CurrentSchools = append(stepState.CurrentSchools, school)

		newStudent, err := s.CreateStudent(stepState.Organization.Staffs[0], gradeMaster, []string{stepState.Organization.DefaultLocation.ID}, true, numParent, school.ID)
		if err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
		stepState.Students = append(stepState.Students, newStudent)
		stepState.GradeAssigneds = append(stepState.GradeAssigneds, gradeMaster)

		for parentIdx := 0; parentIdx < numParent; parentIdx++ {
			stepState.MapStudentIDAndParentIDs[newStudent.ID] = append(stepState.MapStudentIDAndParentIDs[newStudent.ID], newStudent.Parents[parentIdx].ID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CreatesNumberOfStudentsWithSameParentsInfo(ctx context.Context, num string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numStudent := 0
	if num == "random" {
		numStudent = RandRangeIn(2, 5)
	} else {
		var err error
		if numStudent, err = strconv.Atoi(num); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudentsWithSameParentsInfo: %v", err)
		}
	}

	// randomly select a school for student
	idxSchool := RandRangeIn(1, 5)
	school := stepState.Schools[idxSchool]
	stepState.CurrentSchools = append(stepState.CurrentSchools, school)

	grade := RandRangeIn(0, len(stepState.GradeMasters))
	gradeMaster := stepState.GradeMasters[grade]
	opt := &helpers.CreateStudentsWithSameParentOpt{
		NumberOfStudent: numStudent,
	}
	students, _, err := s.CreateStudentsWithSameParent(stepState.Organization.Staffs[0], stepState.CurrentOrganicationID, int32(grade), gradeMaster, []string{stepState.Organization.DefaultLocation.ID}, opt, school.ID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Students = students

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) StudentLoginsToLearnerApp(ctx context.Context, studentIdx string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	switch studentIdx {
	case "all":
		for _, student := range stepState.Students {
			var err error
			student.Token, err = s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	default:
		var (
			studentIndex int
			err          error
		)
		if studentIndex, err = strconv.Atoi(studentIdx); err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfStudentsWithSameParentsInfo: %v", err)
		}
		for i, student := range stepState.Students {
			if i == studentIndex {
				var err error
				student.Token, err = s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

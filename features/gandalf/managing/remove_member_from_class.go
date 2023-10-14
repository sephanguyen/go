package managing

import (
	"context"

	"github.com/manabie-com/backend/features/bob"
)

func (s *suite) joinClassWithSchoolName(ctx context.Context, number int, role, schoolName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	var err error
	for i := 1; i <= number; i++ {
		ctx, err = s.bobSuite.JoinClassWithSchoolName(ctx, 1, role, schoolName)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		if role == "student" {
			stepState.BobStepState.ClassStudentsID = append(stepState.BobStepState.ClassStudentsID, bob.StepStateFromContext(ctx).CurrentStudentID)
		} else if role == "teacher" {
			stepState.BobStepState.ClassOwnersID = append(stepState.BobStepState.ClassOwnersID, bob.StepStateFromContext(ctx).CurrentTeacherID)
		}
	}

	return GandalfStepStateToContext(ctx, stepState), nil
}

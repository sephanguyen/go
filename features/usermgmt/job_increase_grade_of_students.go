package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) generateStudentsWithDifferentGradeAndSchool(ctx context.Context, schoolID int64) error {
	now := time.Now()
	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))
	amountTestStudent := 16
	for i := 0; i <= amountTestStudent; i++ {
		student, err := newStudentEntity()
		if err != nil {
			return errors.Wrap(err, "newStudentEntity")
		}

		if err := multierr.Combine(
			student.PreviousGrade.Set(nil),
			student.ResourcePath.Set(fmt.Sprint(schoolID)),
		); err != nil {
			return err
		}
		_, err = database.Insert(ctx, student, s.BobDBTrace.Exec)
		if err != nil {
			return err
		}
		s.ExistingStudents = append(s.ExistingStudents, student)
		now = now.AddDate(0, 0, -1)
	}
	return nil
}

func (s *suite) listStudentsWithDifferentGradeAndDifferentSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.generateStudentsWithDifferentGradeAndSchool(ctx, constants.JPREPSchool)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.generateStudentsWithDifferentGradeAndSchool(ctx, constants.ManabieSchool)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToIncreaseGradeOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunIncreaseGradeOfStudents(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, time.Now().Format("2006-01-02"), fmt.Sprint(constants.ManabieSchool))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) gradeOfStudentsWasIncreasedByLevel(ctx context.Context, level string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	levelInt, err := strconv.Atoi(level)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentIDs := []string{}
	for _, studentEnt := range s.ExistingStudents {
		studentIDs = append(studentIDs, studentEnt.ID.String)
	}

	query := `
		SELECT student_id, current_grade, previous_grade, resource_path 
		FROM students 
		WHERE student_id = ANY($1)
		ORDER BY resource_path, created_at desc;
	`
	rows, err := s.BobDBTrace.Query(ctx, query, database.TextArray(studentIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	index := 0
	for rows.Next() {
		var studentID, resourcePath string
		var currentGrade, previousGrade *pgtype.Int2
		err := rows.Scan(&studentID, &currentGrade, &previousGrade, &resourcePath)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		studentEnt := s.ExistingStudents[index]
		if resourcePath == fmt.Sprint(constants.JPREPSchool) {
			if previousGrade != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected previous_grade is %v, but actual is %v", studentID, nil, previousGrade.Int)
			}

			if currentGrade.Int != studentEnt.CurrentGrade.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected current_grade is %v, but actual is %v", studentID, studentEnt.CurrentGrade.Int, currentGrade.Int)
			}
		} else {
			if previousGrade.Int != studentEnt.CurrentGrade.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected previous_grade is %v, but actual is %v", studentID, studentEnt.CurrentGrade.Int, previousGrade.Int)
			}

			if resourcePath != studentEnt.ResourcePath.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected resource_path is %v, but actual is %v", studentID, studentEnt.ResourcePath, resourcePath)
			}

			switch previousGrade.Int {
			case 0, 16:
				if currentGrade.Int != studentEnt.CurrentGrade.Int {
					return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected current_grade is %v, but actual is %v", studentID, studentEnt.CurrentGrade.Int, currentGrade.Int)
				}
			default:
				if currentGrade.Int != (studentEnt.CurrentGrade.Int + int16(levelInt)) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("gradeOfStudentsWasIncreasedByLevel err: studentID is %v, expected current_grade is %v, but actual is %v", studentID, studentEnt.CurrentGrade.Int, currentGrade.Int)
				}
			}
		}

		// the s.ExistingStudents and the query has the same index (created_at desc)
		index += 1
	}

	if rows.Err() != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

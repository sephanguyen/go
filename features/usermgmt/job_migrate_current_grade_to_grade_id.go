package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const amountTestGradeMaster = 8

func (s *suite) generateGradeMaster(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	stepState.PartnerInternalIDs = make([]string, 17)
	stepState.GradeInternalIDs = make([]string, 17)
	stepState.ManabieGradeIDs = make([]string, 17)

	stmt := `INSERT INTO grade (grade_id, partner_internal_id, name, updated_at, created_at) VALUES ($1, $2, $3, now(), now())`
	for i := 0; i <= amountTestGradeMaster; i++ {
		gradeID := idutil.ULIDNow()
		partnerID := "partner-" + gradeID
		name := "name-" + gradeID
		cmd, err := s.BobDBTrace.Exec(ctx, stmt, gradeID, partnerID, name)
		if err != nil {
			return ctx, fmt.Errorf("s.BobDBTrace.Exec err: %v", err)
		}
		if cmd.RowsAffected() == 0 {
			return ctx, fmt.Errorf("s.BobDBTrace.Exec err: no row effect")
		}

		stepState.PartnerInternalIDs[i] = partnerID
		stepState.GradeInternalIDs[i] = partnerID
		stepState.ManabieGradeIDs[i] = gradeID
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listStudentsWithGradeMaster(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	s.ExistingStudents = []*entity.LegacyStudent{}
	amountTestStudent := 16
	for i := 0; i <= amountTestStudent; i++ {
		student, err := newStudentEntity()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("newStudentEntity err: %v", err)
		}

		if err := multierr.Combine(
			student.CurrentGrade.Set(i),
			student.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine err: %v", err)
		}
		_, err = database.Insert(ctx, student, s.BobDBTrace.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("database.Insert err: %v", err)
		}
		s.ExistingStudents = append(s.ExistingStudents, student)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateCurrentGradeToGradeID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunMigrateCurrentGradeToGradeID(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, strings.Join(stepState.PartnerInternalIDs, "|"), fmt.Sprint(constants.ManabieSchool))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) gradeOfStudentsWasMigratedToGradeID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentIDs := []string{}
	for _, studentEnt := range s.ExistingStudents {
		studentIDs = append(studentIDs, studentEnt.ID.String)
	}

	gradeIDs := []string{}
	for _, partnerID := range stepState.PartnerInternalIDs {
		if partnerID == "" {
			continue
		}

		gradeIDs = append(gradeIDs, strings.Split(partnerID, "-")[1])
	}

	query := `
		SELECT s.student_id, s.grade_id, s.resource_path 
		FROM students s
		LEFT JOIN grade_organization go
		ON s.current_grade = go.grade_value AND s.grade_id = go.grade_id
		WHERE student_id = ANY($1)
		ORDER BY s.current_grade;
	`
	rows, err := s.BobDBTrace.Query(ctx, query, database.TextArray(studentIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	index := 0
	for rows.Next() {
		var studentID, resourcePath string
		var gradeID pgtype.Text
		err := rows.Scan(&studentID, &gradeID, &resourcePath)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		studentEnt := s.ExistingStudents[index]
		if studentID != studentEnt.ID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("validation err: expected studentID is %v, but actual is %v", studentEnt.ID.String, studentID)
		}
		if index > amountTestGradeMaster {
			if gradeID.String != "" {
				return StepStateToContext(ctx, stepState), fmt.Errorf(`validation err: studentID is %v, expected gradeID is "", but actual is %v`, studentID, gradeID)
			}
		} else {
			if !golibs.InArrayString(gradeID.String, gradeIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validation err: studentID is %v, expected gradeID is %v, but actual is %v", studentID, gradeID, gradeIDs)
			}
		}

		if resourcePath != studentEnt.ResourcePath.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("validation err: studentID is %v, expected resource_path is %v, but actual is %v", studentID, studentEnt.ResourcePath, resourcePath)
		}
		index++
	}

	if rows.Err() != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

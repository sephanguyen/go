package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
)

var currentSchoolScanQuery = `SELECT t1.school_id FROM school_history t1
    INNER JOIN students t6 on t1.student_id = t6.student_id
  	INNER JOIN school_info t2 ON t1.school_id = t2.school_id
  	INNER JOIN school_level t3 ON t2.school_level_id = t3.school_level_id
  	INNER JOIN school_level_grade t4 ON t3.school_level_id = t4.school_level_id
  	INNER JOIN grade t5 on t4.grade_id = t5.grade_id
	WHERE t1.is_current = false AND t1.deleted_at IS NULL
	LIMIT $1;
	`

func (s *suite) systemRunJobToMigrateSetCurrentSchoolByGradeInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunMigrateSetCurrentSchoolByGrade(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingSchoolHistoryWithCurrentSchoolValueSetByGradeValue(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rows, err := s.BobPostgresDBTrace.Query(
		ctx,
		currentSchoolScanQuery,
		limit,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("existingSchoolHistoryWithCurrentSchoolValueSetByGradeValue: query error %s", err.Error())
	}
	defer rows.Close()

	schoolIDs := []string{}

	for rows.Next() {
		schoolID := ""
		if err := rows.Scan(&schoolID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		schoolIDs = append(schoolIDs, schoolID)
	}

	if len(schoolIDs) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("migrate set current school by grade fail")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateSchoolHistoryWithoutCurrentSchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE school_history SET is_current = false WHERE deleted_at IS NULL`
	_, err := s.BobDBTrace.Exec(ctx, stmt)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

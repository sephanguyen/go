package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userChoosesAStudyPlanInCourseForDeleting(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT 
			csp.course_id,
			csp.study_plan_id
		FROM course_study_plans as csp
		JOIN study_plans AS sp ON sp.study_plan_id = csp.study_plan_id
		WHERE sp.master_study_plan_id IS NULL
    		AND sp.deleted_at IS NULL
			AND csp.deleted_at IS NULL
		LIMIT 1 
	`
	if err := s.DB.QueryRow(ctx, query).Scan(stepState.CourseID, stepState.StudyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	query2 := `
		SELECT 
			ssp.student_id
		FROM student_study_plans as ssp
		JOIN study_plans AS sp ON sp.study_plan_id = ssp.study_plan_id
		WHERE sp.master_study_plan_id  = $1
			AND sp.deleted_at IS NULL
			AND ssp.deleted_at IS NULL
		LIMIT 1 
	`
	if err := s.DB.QueryRow(ctx, query2, stepState.StudyPlanID).Scan(stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeletesSelectedStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken
	req := &epb.DeleteStudyPlanBelongsToACourseRequest{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
	}

	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanModifierServiceClient(s.Conn).DeleteStudyPlanBelongsToACourse(s.signedCtx(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.IsDeleteStudyPlan = true

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) fetchsNewStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (retry bool, err error) {
		if stepState.CourseID == "" && stepState.StudyPlanID == "" && stepState.StudentID == "" {
			time.Sleep(500 * time.Millisecond)
			return attempt < 10, fmt.Errorf("need course_id, study_plan_id and student_id from before scenario")
		} else {
			return false, nil
		}
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = try.Do(func(attempt int) (retry bool, err error) {
		defer func() {
			if retry {
				time.Sleep(500 * time.Millisecond)
			}
		}()

		stepState.AuthToken = stepState.TeacherToken
		pbStudyPlans := make([]*epb.StudyPlan, 0)
		paging := &cpb.Paging{
			Limit: 5,
		}
		res1, err := epb.NewStudyPlanReaderServiceClient(s.Conn).ListStudyPlanByCourse(s.signedCtx(ctx), &epb.ListStudyPlanByCourseRequest{
			CourseId: stepState.CourseID,
			Paging:   paging,
		})
		if err != nil {
			return attempt < 10, err
		}
		pbStudyPlans = append(pbStudyPlans, res1.StudyPlans...)

		if len(pbStudyPlans) == len(stepState.TeacherStudyPlans) {
			return attempt < 10, fmt.Errorf("this deleting function does'nt work")
		}

		for _, item := range pbStudyPlans {
			if item.StudyPlanId == stepState.StudyPlanID {
				return attempt < 10, fmt.Errorf("study plan isn't deleted")
			}
		}

		stepState.AuthToken = stepState.StudentToken
		res2, err := epb.NewAssignmentReaderServiceClient(s.Conn).ListStudentAvailableContents(s.signedCtx(ctx), &epb.ListStudentAvailableContentsRequest{
			CourseId: stepState.CourseID,
		})
		if err != nil {
			return attempt < 10, err
		}

		var count int
		query2 := `
		SELECT 
			COUNT(*) as count
		FROM student_study_plans as ssp
		WHERE ssp.student_id  = $1
		AND ssp.deleted_at IS NULL
		LIMIT 1 
	`
		if err := s.DB.QueryRow(s.signedCtx(ctx), query2, &stepState.StudentID).Scan(&count); err != nil {
			return false, err
		}
		if count != len(res2.Contents) {
			return attempt < 10, fmt.Errorf("student fecths errors")
		}
		return false, nil
	})
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) fetchsOldStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.TeacherToken
	pbStudyPlans := make([]*epb.StudyPlan, 0)
	paging := &cpb.Paging{
		Limit: 5,
	}
	res, err := epb.NewStudyPlanReaderServiceClient(s.Conn).ListStudyPlanByCourse(s.signedCtx(ctx), &epb.ListStudyPlanByCourseRequest{
		CourseId: stepState.CourseID,
		Paging:   paging,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	pbStudyPlans = append(pbStudyPlans, res.StudyPlans...)

	stepState.TeacherStudyPlans = pbStudyPlans

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkSelectedStudyPlanHasBeenAbsolutelyDeleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (retry bool, err error) {
		if !stepState.IsDeleteStudyPlan {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("study plan must be deleted in before scenario")
		}
		return false, nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var studyPlanCount int
	var studyPlanItemCount int
	var studyPlanCourseCount int
	var studyPlanStudentCount int

	recursiveQuery := `
		WITH RECURSIVE study_plan_recurs (study_plan_id, master_study_plan_id, deleted_at) AS (
			SELECT sp1.study_plan_id,
						sp1.master_study_plan_id,
						sp1.deleted_at
			FROM study_plans sp1
			WHERE sp1.study_plan_id = $1
				AND sp1.master_study_plan_id IS NULL

			UNION ALL

			SELECT sp2.study_plan_id,
						sp2.master_study_plan_id,
						sp2.deleted_at
			FROM study_plans as sp2
							JOIN study_plan_recurs spr ON spr.study_plan_id = sp2.master_study_plan_id
		)
	`

	selectStudyPlanQuery := recursiveQuery + ` SELECT COUNT(*) FROM study_plans WHERE study_plan_id IN (SELECT spr.study_plan_id FROM study_plan_recurs AS spr) and deleted_at IS NULL`
	if err := s.DB.QueryRow(ctx, selectStudyPlanQuery, stepState.StudyPlanID).Scan(&studyPlanCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studyPlanCount != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("[study plans] this function not delete study plan")
	}

	selectStudyPlanItemQuery := recursiveQuery + ` 
		SELECT COUNT(*) FROM (
			SELECT * FROM (SELECT sp.study_plan_id FROM study_plans AS sp WHERE sp.study_plan_id IN (SELECT spr.study_plan_id FROM study_plan_recurs AS spr)) AS temp
				JOIN study_plan_items as spi
					ON temp.study_plan_id = spi.study_plan_id AND spi.deleted_at IS NULL
		) as noname
	`
	if err = s.DB.QueryRow(ctx, selectStudyPlanItemQuery, stepState.StudyPlanID).Scan(&studyPlanItemCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studyPlanItemCount != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("[study plan items] this function not delete study plan")
	}

	selectStudyPlanCourseQuery := `SELECT COUNT(*) FROM course_study_plans WHERE course_id = $1 AND deleted_at IS NULL`
	if err = s.DB.QueryRow(ctx, selectStudyPlanCourseQuery, stepState.CourseID).Scan(&studyPlanCourseCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studyPlanCourseCount != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("[course study plans]: this function not delete study plan")
	}

	selectStudentStudyPlanQuery := recursiveQuery + ` 
		SELECT COUNT(*) FROM (
			SELECT * FROM (SELECT sp.study_plan_id FROM study_plans AS sp WHERE sp.study_plan_id IN (SELECT spr.study_plan_id FROM study_plan_recurs AS spr)) AS temp
				JOIN student_study_plans as ssp
					ON temp.study_plan_id = ssp.study_plan_id AND ssp.deleted_at IS NULL
		) as noname
	`
	if err := s.DB.QueryRow(ctx, selectStudentStudyPlanQuery, stepState.StudyPlanID).Scan(&studyPlanStudentCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if studyPlanStudentCount != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("[student study plans]: this function not delete study plan")
	}

	return StepStateToContext(ctx, stepState), nil
}

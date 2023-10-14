package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) generateStudyPlan(ctx context.Context, masterStudyPlanID string, index string) (*entities.StudyPlan, error) {
	stepState := StepStateFromContext(ctx)
	id := idutil.ULIDNow()
	if stepState.CourseID == "" {
		stepState.CourseID = idutil.ULIDNow()
	}
	now := timeutil.Now()
	e := &entities.StudyPlan{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.ID.Set(id),
		e.Name.Set(fmt.Sprintf("name_%s", id)),
		e.SchoolID.Set(constants.ManabieSchool),
		e.CourseID.Set(stepState.CourseID),
		e.Status.Set(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
		e.BaseEntity.CreatedAt.Set(now),
		e.BaseEntity.UpdatedAt.Set(now),
	)
	if masterStudyPlanID != "" {
		err = multierr.Append(err, e.MasterStudyPlan.Set(masterStudyPlanID))
	}
	if index != "" {
		err = multierr.Append(err, e.Name.Set(fmt.Sprintf("name_%s_%s", id, index)))
	}
	if err != nil {
		return nil, fmt.Errorf("unable to set value to entity StudyPlan: %w", err)
	}

	return e, nil
}

func (s *suite) aCourseStudyPlansCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanRepo := &repositories.StudyPlanRepo{}
	e, err := s.generateStudyPlan(ctx, "", "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CourseStudyPlanID = e.ID.String
	_, err = studyPlanRepo.Insert(ctx, s.DB, e)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert study plan: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourDatabaseHaveToRaiseAnErrorViolateUniqueIndex(ctx context.Context, caseStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if caseStatus == "duplicate" {
		if stepState.ResponseErr.Error() != `batchResults.Exec: ERROR: duplicate key value violates unique constraint "student_master_study_plan"` {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) genStudentStudyPlan(ctx context.Context, studyPlanID string) (context.Context, *entities.StudentStudyPlan) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudentStudyPlan{}
	database.AllNullEntity(e)
	if stepState.StudentID == "" {
		stepState.StudentID = idutil.ULIDNow()
	}
	now := timeutil.Now()
	_ = multierr.Combine(
		e.StudentID.Set(stepState.StudentID),
		e.StudyPlanID.Set(studyPlanID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now))

	return StepStateToContext(ctx, stepState), e
}

func (s *suite) theDatabaseAdminCreateSomeStudentStudyPlans(ctx context.Context, statusCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomNum := rand.Intn(10) + 2
	if statusCase == "no duplicate" {
		randomNum = 1
	}
	studyPlanArr := make([]*entities.StudyPlan, 0, randomNum)
	studentStudyPlanArr := make([]*entities.StudentStudyPlan, 0, randomNum)
	for i := 0; i < randomNum; i++ {
		studyPlanEnt, err := s.generateStudyPlan(ctx, stepState.CourseStudyPlanID, strconv.Itoa(i+1))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, studentStudyPlanEnt := s.genStudentStudyPlan(ctx, studyPlanEnt.ID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyPlanArr = append(studyPlanArr, studyPlanEnt)
		studentStudyPlanArr = append(studentStudyPlanArr, studentStudyPlanEnt)
	}
	studyPlanRepo := &repositories.StudyPlanRepo{}
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := studyPlanRepo.BulkUpsert(ctx, tx, studyPlanArr)
		if err != nil {
			return fmt.Errorf("unable to upsert study plan: %w", err)
		}
		err = studentStudyPlanRepo.BulkUpsert(ctx, tx, studentStudyPlanArr)
		if err != nil {
			return fmt.Errorf("unable to upsert student study plan: %w", err)
		}

		return nil
	})
	if e, isErr := errors.Cause(err).(*pgconn.PgError); isErr && e.Code != "23505" {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

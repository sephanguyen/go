package task_assignment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) userCreatesAValidAdhocTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp, err := sspb.NewTaskAssignmentClient(s.EurekaConn).UpsertAdhocTaskAssignment(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&sspb.UpsertAdhocTaskAssignmentRequest{
			StudentId:      stepState.UserID,
			CourseId:       stepState.CourseID,
			StartDate:      timestamppb.New(time.Now()),
			TaskAssignment: &sspb.TaskAssignmentBase{},
		},
	)
	stepState.Response = resp
	stepState.ResponseErr = err
	if err == nil {
		stepState.LearningMaterialID = resp.LearningMaterialId
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemCreatesAdhocTaskAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lmID := stepState.LearningMaterialID
	e := entities.LearningMaterial{}
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE learning_material_id = $1", e.TableName())
	var c int
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(lmID)).Scan(&c); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}
	if c == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ad hoc task assignment with id = %s not found", lmID)
	}

	t, _ := time.Parse("2006/02/01 15:04", "2300/01/01 23:59")

	me := &entities.MasterStudyPlan{}
	var availableTo pgtype.Timestamptz
	query = fmt.Sprintf("SELECT available_to FROM %s WHERE learning_material_id = $1", me.TableName())
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(lmID)).Scan(&availableTo); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}
	if !availableTo.Time.Equal(t) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("available to of master study plan is wrong")
	}
	ie := &entities.IndividualStudyPlan{}
	query = fmt.Sprintf("SELECT available_to FROM %s WHERE learning_material_id = $1", ie.TableName())
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(lmID)).Scan(&availableTo); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}
	if !availableTo.Time.Equal(t) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("available to of individual study plan is wrong")
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdatesTheAdhocTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if _, err := sspb.NewTaskAssignmentClient(s.EurekaConn).UpsertAdhocTaskAssignment(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&sspb.UpsertAdhocTaskAssignmentRequest{
			StudentId: stepState.UserID,
			CourseId:  stepState.CourseID,
			StartDate: timestamppb.New(time.Now()),
			TaskAssignment: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					LearningMaterialId: stepState.LearningMaterialID,
				},
			},
		},
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.NewTaskAssignmentClient(s.EurekaConn).UpsertAdhocTaskAssignment: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesAdhocTaskAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lmID := stepState.LearningMaterialID
	e := entities.LearningMaterial{}
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE learning_material_id = $1 AND created_at <> updated_at", e.TableName())
	var c int
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(lmID)).Scan(&c); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.EurekaDB.QueryRow: %w", err)
	}
	if c == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ad hoc task assignment with id = %s not updated", lmID)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

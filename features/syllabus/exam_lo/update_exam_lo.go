package exam_lo

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	examLOUpdatedName        = "ExamLO updated"
	examLOUpdatedInstruction = "Updated instruction"
)

func (s *Suite) userUpdateAValidExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.UpdateExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: stepState.LearningMaterialIDs[0],
				Name:               examLOUpdatedName,
			},
			Instruction: examLOUpdatedInstruction,
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).UpdateExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdateExamLOCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ex := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.LearningMaterialIDs[0]),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(ex), ","), ex.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, ex.ID).ScanOne(ex); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if ex.Name.String != examLOUpdatedName {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect ExamLO name: expected %s, got %s", examLOUpdatedName, ex.Name.String)
	}
	if ex.Instruction.String != examLOUpdatedInstruction {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect ExamLO Instruction: expected %s, got %s", examLOUpdatedName, ex.Instruction.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateAExamLOWithField(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.UpdateExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				Name:               "Exam LO",
				LearningMaterialId: stepState.LearningMaterialIDs[0],
			},
		},
	}
	n := rand.Intn(5) + 3
	switch field {
	case "instruction":
		req.ExamLo.Instruction = "updated instruction"
	case "manual_grading":
		req.ExamLo.ManualGrading = true
	case "time_limit":
		req.ExamLo.TimeLimit = wrapperspb.Int32(int32(n))
	case "grade_to_pass":
		req.ExamLo.GradeToPass = wrapperspb.Int32(int32(n))
	case "maximum_attempt":
		req.ExamLo.MaximumAttempt = wrapperspb.Int32(int32(n + 1))
	case "approve_grading":
		req.ExamLo.ApproveGrading = n%2 == 0
	case "grade_capping":
		req.ExamLo.GradeCapping = n%2 != 0
	case "review_option":
		req.ExamLo.ReviewOption = sspb.ExamLOReviewOption(int32(math.Min(1, float64(n%2))))
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to exam LO", field)
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).UpdateExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustUpdateExamLOWithUpdatedFieldCorrectly(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	updateReq := stepState.Request.(*sspb.UpdateExamLORequest)

	stepState.ExamLOBase = updateReq.ExamLo
	LearningMaterialIds := []string{updateReq.ExamLo.Base.LearningMaterialId}

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListExamLORequest{
		LearningMaterialIds: LearningMaterialIds,
	})
	listExamLOResp := stepState.Response.(*sspb.ListExamLOResponse)
	for _, exam := range listExamLOResp.ExamLos {
		switch field {
		case "instruction":
			if exam.Instruction != stepState.ExamLOBase.Instruction {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO instruction: want %s, got %s", stepState.ExamLOBase.Instruction, exam.Instruction)
			}
		case "manual_grading":
			if exam.ManualGrading != stepState.ExamLOBase.ManualGrading {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO manual_grading: want %t, got %t", stepState.ExamLOBase.ManualGrading, exam.ManualGrading)
			}
		case "time_limit":
			if exam.TimeLimit.String() != stepState.ExamLOBase.TimeLimit.String() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO time_limit: want %s, got %s", stepState.ExamLOBase.TimeLimit.String(), exam.TimeLimit.String())
			}
		case "grade_to_pass":
			if exam.GradeToPass.String() != stepState.ExamLOBase.GradeToPass.String() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO grade_to_pass: want %s, got %s", stepState.ExamLOBase.GradeToPass.String(), exam.GradeToPass.String())
			}
		case "maximum_attempt":
			if exam.MaximumAttempt.GetValue() != stepState.ExamLOBase.MaximumAttempt.GetValue() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO maximum_attempt: want %d, got %d", stepState.ExamLOBase.MaximumAttempt.GetValue(), exam.MaximumAttempt.GetValue())
			}
		case "approve_grading":
			if exam.ApproveGrading != stepState.ExamLOBase.ApproveGrading {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO approve_grading: want %t, got %t", stepState.ExamLOBase.ApproveGrading, exam.ApproveGrading)
			}
		case "grade_capping":
			if exam.GradeCapping != stepState.ExamLOBase.GradeCapping {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO grade_capping: want %t, got %t", stepState.ExamLOBase.GradeCapping, exam.GradeCapping)
			}
		case "review_option":
			if exam.ReviewOption.String() != stepState.ExamLOBase.ReviewOption.String() {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam LO review_option: want %s, got %s", stepState.ExamLOBase.ReviewOption.String(), exam.ReviewOption.String())
			}
		default:
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to exam LO", field)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

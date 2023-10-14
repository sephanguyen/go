package learning_material

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) ourSystemMustReturnLearningMaterialCorrectly(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	switch arg1 {
	case "assignment":
		response := stepState.Response.(*sspb.ListLearningMaterialResponse).GetAssignment()
		assignments := response.Assignments
		if len(assignments) < 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of assignments, expected %d, got %d", len(stepState.GeneralAssignmentIDs), len(assignments))
		}
		for _, assignment := range assignments {
			if !golibs.InArrayString(assignment.Base.LearningMaterialId, stepState.GeneralAssignmentIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected assignment id: %q in list %v of assignment: %q", assignment.Base.LearningMaterialId, stepState.GeneralAssignmentIDs, assignments)
			}
		}
	case "learning_objective":
		response := stepState.Response.(*sspb.ListLearningMaterialResponse).GetLearningObjective()
		los := response.LearningObjectives
		if len(los) < 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of learning objectives, expected %d, got %d", len(stepState.LearningObjectiveIDs), len(los))
		}
		for _, lo := range los {
			if !golibs.InArrayString(lo.Base.LearningMaterialId, stepState.LearningObjectiveIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective id: %q in list %v of learning objectives: %q", lo.Base.LearningMaterialId, stepState.LearningObjectiveIDs, los)
			}
		}
	case "flashcard":
		response := stepState.Response.(*sspb.ListLearningMaterialResponse).GetFlashcard()
		flashcards := response.Flashcards
		if len(flashcards) < 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of flashcards, expected %d, got %d", len(stepState.FlashcardIDs), len(flashcards))
		}
		for _, flashcard := range flashcards {
			if !golibs.InArrayString(flashcard.Base.LearningMaterialId, stepState.FlashcardIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected flashcard id: %q in list %v of flashcards: %q", flashcard.Base.LearningMaterialId, stepState.FlashcardIDs, flashcards)
			}
		}
	case "exam_lo":
		response := stepState.Response.(*sspb.ListLearningMaterialResponse).GetExamLo()
		examLOs := response.ExamLos
		if len(examLOs) < 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of exam los, expected %d, got %d", len(stepState.ExamLOIDs), len(examLOs))
		}
		for _, examLo := range examLOs {
			if !golibs.InArrayString(examLo.Base.LearningMaterialId, stepState.ExamLOIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected exam lo id: %q in list %v of exam lo: %q", examLo.Base.LearningMaterialId, stepState.ExamLOIDs, examLOs)
			}
		}
	case "task_assignment":
		response := stepState.Response.(*sspb.ListLearningMaterialResponse).GetTaskAssignment()
		taskAssignments := response.TaskAssignments
		if len(taskAssignments) < 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of task assignment, expected %d, got %d", len(stepState.TaskAssignmentIDs), len(taskAssignments))
		}
		for _, taskAssignment := range taskAssignments {
			if !golibs.InArrayString(taskAssignment.Base.LearningMaterialId, stepState.TaskAssignmentIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected task assignment id: %q in list %v of task assignment: %q", taskAssignment.Base.LearningMaterialId, stepState.TaskAssignmentIDs, taskAssignments)
			}
		}

	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("type mismatch in request and return checking function")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userSendListLearningMaterialRequest(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	switch arg1 {
	case "assignment":
		stepState.LearningMaterialID = stepState.MapLearningMaterial[arg1].LearningMaterialBase.GetLearningMaterialId()
		stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).ListLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningMaterialRequest{
			Message: &sspb.ListLearningMaterialRequest_Assignment{
				Assignment: &sspb.ListAssignmentRequest{
					LearningMaterialIds: stepState.GeneralAssignmentIDs,
				},
			},
		})
	case "learning_objective":
		stepState.LearningMaterialID = stepState.MapLearningMaterial[arg1].LearningMaterialBase.GetLearningMaterialId()
		stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).ListLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningMaterialRequest{
			Message: &sspb.ListLearningMaterialRequest_LearningObjective{
				LearningObjective: &sspb.ListLearningObjectiveRequest{
					LearningMaterialIds: stepState.LearningObjectiveIDs,
				},
			},
		})
	case "flashcard":
		stepState.LearningMaterialID = stepState.MapLearningMaterial[arg1].LearningMaterialBase.GetLearningMaterialId()
		stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).ListLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningMaterialRequest{
			Message: &sspb.ListLearningMaterialRequest_Flashcard{
				Flashcard: &sspb.ListFlashcardRequest{
					LearningMaterialIds: stepState.FlashcardIDs,
				},
			},
		})
	case "exam_lo":
		stepState.LearningMaterialID = stepState.MapLearningMaterial[arg1].LearningMaterialBase.GetLearningMaterialId()
		stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).ListLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningMaterialRequest{
			Message: &sspb.ListLearningMaterialRequest_ExamLo{
				ExamLo: &sspb.ListExamLORequest{
					LearningMaterialIds: stepState.ExamLOIDs,
				},
			},
		})
	case "task_assignment":
		stepState.LearningMaterialID = stepState.MapLearningMaterial[arg1].LearningMaterialBase.GetLearningMaterialId()
		stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).ListLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListLearningMaterialRequest{
			Message: &sspb.ListLearningMaterialRequest_TaskAssignment{
				TaskAssignment: &sspb.ListTaskAssignmentRequest{
					LearningMaterialIds: stepState.TaskAssignmentIDs,
				},
			},
		})

	default:
		return utils.StepStateToContext(ctx, stepState), nil
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userSendListArbitraryLearningMaterialRequest(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	lmTypes := []string{"assignment", "learning_objective", "flashcard", "exam_lo", "task_assignment"}
	randomN := rand.Intn(len(lmTypes))
	ctx, err := s.userSendListLearningMaterialRequest(ctx, lmTypes[randomN])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to send list arbitrary learning material request, err: %v", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

package learning_material

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) aValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookID, chapterIDs, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, s.EurekaDB, constant.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeletesTheLearningMaterial(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LearningMaterialID = stepState.MapLearningMaterial[arg].LearningMaterialBase.GetLearningMaterialId()

	if _, err := sspb.NewLearningMaterialClient(s.EurekaConn).DeleteLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DeleteLearningMaterialRequest{
		LearningMaterialId: stepState.LearningMaterialID,
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete the learning material: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint: gosec
func (s *Suite) userDeletesAnArbitraryTheLearningMaterial(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	//TODO: add the rest type.
	lmTypes := []string{"assignment", "learning_objective", "flashcard"}
	randomN := rand.Intn(len(lmTypes))
	stepState.LearningMaterialID = stepState.MapLearningMaterial[lmTypes[randomN]].LearningMaterialBase.GetLearningMaterialId()
	_, err := sspb.NewLearningMaterialClient(s.EurekaConn).DeleteLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DeleteLearningMaterialRequest{
		LearningMaterialId: stepState.LearningMaterialID,
	})
	stepState.ResponseErr = err
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustDeleteTheLearningMaterialCorrectly(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	expectedCount := 1

	e := entities.LearningMaterial{}
	if _, err := s.validateLMDeleted(ctx, e.TableName(), expectedCount); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	switch arg {
	case "assignment":
		e := entities.GeneralAssignment{}
		if _, err := s.validateLMDeleted(ctx, e.TableName(), expectedCount); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case "learning_objective":
		e := entities.LearningObjectiveV2{}
		if _, err := s.validateLMDeleted(ctx, e.TableName(), expectedCount); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case "flashcard":
		e := entities.Flashcard{}
		if _, err := s.validateLMDeleted(ctx, e.TableName(), expectedCount); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case "task_assignment":
		// TODO
	case "exam_lo":
		// TODO
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint: unparam
func (s *Suite) validateLMDeleted(ctx context.Context, tableName string, expectedCount int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// this should be pgtype
	var deletedLMCount int

	rawQuery := `SELECT COUNT(*) FROM %s WHERE learning_material_id = $1::TEXT AND deleted_at IS NOT NULL`
	query := fmt.Sprintf(rawQuery, tableName)
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(stepState.LearningMaterialID)).Scan(&deletedLMCount); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("failed to query table %s count: %w", tableName, err)
	}
	if deletedLMCount != expectedCount {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("failed to deleted  on table %s expected: %v, got: %v", tableName, expectedCount, deletedLMCount)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeletesTheLearningMaterialWithMissingID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).DeleteLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DeleteLearningMaterialRequest{
		LearningMaterialId: "",
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeletesTheLearningMaterialWithWrongID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).DeleteLearningMaterial(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DeleteLearningMaterialRequest{
		LearningMaterialId: "wrong ID",
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

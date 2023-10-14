package shuffled_quiz_set

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) validQuizSet(ctx context.Context, arg int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err1 := s.aSignedIn(ctx, "school admin")
	ctx, err2 := s.aValidBookContent(ctx)
	ctx, err3 := s.userCreateACourseWithAStudyPlan(ctx)
	ctx, err4 := s.userCreateAQuizUsingV2(ctx)
	if err := multierr.Combine(err1, err2, err3, err4); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("validQuizSet: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentDoQuizTestSuccess(ctx context.Context, arg int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for i := 0; i < arg; i++ {
		ctx, err1 := s.aSignedIn(ctx, "student")
		ctx, err2 := s.userCreateQuizTestV2(ctx)
		if err := multierr.Combine(err1, err2); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("createAQuizV2[%d]: %w", i, err)
		}

		stepState.StudyPlanItemIdentities = append(stepState.StudyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoID,
			StudentId:          wrapperspb.String(stepState.Student.ID),
		})
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) teacherGetQuizTestByStudyPlanItemIdentity(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return s.retrieveQuizTestsV2(ctx, "teacher", stepState.StudyPlanItemIdentities)
}

func (s *Suite) teacherGetQuizTestWithoutStudyPlanItemIdentity(ctx context.Context) (context.Context, error) {
	return s.retrieveQuizTestsV2(ctx, "teacher", nil)
}

func (s *Suite) quizTestsInfo(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	res, ok := stepState.Response.(*sspb.RetrieveQuizTestV2Response)
	if !ok {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("stepState.Response.(*sspb.RetrieveQuizTestV2Response) incorrect")
	}

	numTests := arg1
	if numTests != len(res.Items) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("returns num of quiz tests expect %v but got %v", numTests, len(res.Items))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) retrieveQuizTestsV2(ctx context.Context, role string, identities []*sspb.StudyPlanItemIdentity) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aSignedIn(ctx, role)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("aSignedIn %w", err)
	}

	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).RetrieveQuizTestsV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.RetrieveQuizTestV2Request{
		StudyPlanItemIdentities: identities,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) retrieveQuizTestsV2ByRole(ctx context.Context, role string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return s.retrieveQuizTestsV2(ctx, role, stepState.StudyPlanItemIdentities)
}

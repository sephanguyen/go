package shuffled_quiz_set

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreateRetryQuizTestV(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.CreateQuizTestV2Response)
	stepState.ShuffledQuizSetID = resp.ShuffleQuizSetId
	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).CreateRetryQuizTestV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.CreateRetryQuizTestV2Request{

		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LoID,
			StudentId:          wrapperspb.String(stepState.Student.ID),
		},
		ShuffleQuizSetId: wrapperspb.String(stepState.ShuffledQuizSetID),
		SessionId:        idutil.ULIDNow(),
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) retryShuffledQuizTestHaveBeenStored(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.CreateRetryQuizTestV2Response)
	stepState.RetryShuffledQuizSetID = resp.ShuffleQuizSetId
	repo := &repositories.ShuffledQuizSetRepo{}
	if _, err := repo.Get(ctx, s.EurekaDB, database.Text(stepState.RetryShuffledQuizSetID), database.Int8(1), database.Int8(1)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve retry shuffle quiz test: %v", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

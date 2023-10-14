package common

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *NotificationSuite) CurrentStaffCreateQuestionnaire(ctx context.Context, resubmit, questions string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isAllowResubmit := false
	if resubmit == "true" {
		isAllowResubmit = true
	}

	questionnaire := &cpb.Questionnaire{
		QuestionnaireId: idutil.ULIDNow(),
		ResubmitAllowed: isAllowResubmit,
		Questions:       s.CommunicationHelper.ParseQuestionFromString(questions),
		ExpirationDate:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	stepState.Questionnaire = questionnaire
	return StepStateToContext(ctx, stepState), nil
}

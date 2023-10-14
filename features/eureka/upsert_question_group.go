package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *suite) insertQuestionGroupWithRichDescription(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &sspb.UpsertQuestionGroupRequest{
		LearningMaterialId: stepState.LoID,
		Name:               "name",
		Description:        "description",
		RichDescription: &cpb.RichText{
			Raw:      "raw rich text",
			Rendered: "rendered rich text",
		},
	}
	switch arg {
	case "empty":
		req.RichDescription.Raw = ""
		req.RichDescription.Rendered = ""
	case "null":
		req.RichDescription = nil
	case "full":
		break
	}

	res, err := s.upsertQuestionGroup(ctx, req)
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = res, err
	if err == nil {
		stepState.ExistingQuestionHierarchy.AddQuestionGroupID(res.QuestionGroupId)
		stepState.QuestionGroupID = res.QuestionGroupId
		stepState.QuestionGroupIDs = append(stepState.QuestionGroupIDs, res.QuestionGroupId)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertANewQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &sspb.UpsertQuestionGroupRequest{
		LearningMaterialId: stepState.LoID,
		Name:               "name",
		Description:        "description",
		RichDescription: &cpb.RichText{
			Raw:      "raw rich text",
			Rendered: "rendered rich text",
		},
	}
	res, err := s.upsertQuestionGroup(ctx, req)
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = res, err
	if err == nil {
		stepState.ExistingQuestionHierarchy.AddQuestionGroupID(res.QuestionGroupId)
		stepState.QuestionGroupID = res.QuestionGroupId
		stepState.QuestionGroupIDs = append(stepState.QuestionGroupIDs, res.QuestionGroupId)
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) upsertQuestionGroup(ctx context.Context, req *sspb.UpsertQuestionGroupRequest) (*sspb.UpsertQuestionGroupResponse, error) {
	if len(req.LearningMaterialId) == 0 {
		return nil, fmt.Errorf("lo ID dont have yet")
	}
	ctx = s.signedCtx(ctx)
	return sspb.NewQuestionServiceClient(s.Conn).
		UpsertQuestionGroup(ctx, req)
}

func (s *suite) existingQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := stepState.AuthToken
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	ctx, err = s.insertANewQuestionGroup(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingQuestionGroups(ctx context.Context, numItems int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := stepState.AuthToken
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i := 0; i < numItems; i++ {
		ctx, err = s.insertANewQuestionGroup(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) questionGroupReturnedInResp(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.AnswerLogs) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	resp := stepState.Response.(*epb.ListQuizzesOfLOResponse)

	respQuestionGroupIDs := []string{}
	for _, group := range resp.QuestionGroups {
		respQuestionGroupIDs = append(respQuestionGroupIDs, group.QuestionGroupId)
	}

	if !sliceutils.UnorderedEqual(respQuestionGroupIDs, stepState.QuestionGroupIDs) {
		return ctx, fmt.Errorf("question group is not existed in resp, expect %v got %v", respQuestionGroupIDs, stepState.QuestionGroupIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}

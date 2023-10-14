package learning_objective

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) insertANewQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.UpsertQuestionGroupRequest{
		LearningMaterialId: stepState.LearningObjectiveID,
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
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newQuestionGroupWasAddedAtTheEndList(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if stepState.ResponseErr != nil {
		return ctx, nil
	}

	// check new question group data
	repo := &repositories.QuestionGroupRepo{}
	res, err := repo.FindByID(ctx, s.EurekaDB, stepState.QuestionGroupID)
	if err != nil {
		return ctx, fmt.Errorf("QuestionGroupRepo.FindByID: %w", err)
	}

	expected := stepState.Request.(*sspb.UpsertQuestionGroupRequest)
	if res.QuestionGroupID.String != stepState.QuestionGroupID {
		return ctx, fmt.Errorf("expected question group id is %s but got %s", stepState.QuestionGroupID, res.QuestionGroupID.String)
	}
	if res.LearningMaterialID.String != expected.LearningMaterialId {
		return ctx, fmt.Errorf("expected learning material id is %s but got %s", expected.LearningMaterialId, res.LearningMaterialID.String)
	}
	if res.Name.String != expected.Name {
		return ctx, fmt.Errorf("expected name is %s but got %s", expected.Name, res.Name.String)
	}
	if res.Description.String != expected.Description {
		return ctx, fmt.Errorf("expected description is %s but got %s", expected.Description, res.Description.String)
	}

	var richDescription entities.RichText
	err = res.RichDescription.AssignTo(&richDescription)
	if err != nil {
		return ctx, err
	}

	if richDescription.Raw != expected.RichDescription.Raw {
		return ctx, fmt.Errorf("expected raw description is %s but got %s", expected.RichDescription.Raw, richDescription.Raw)
	}

	url, _ := s3.GenerateUploadURL("", "", expected.RichDescription.Rendered)
	if !strings.Contains(richDescription.RenderedURL, strings.ReplaceAll(url, "//", "")) {
		return ctx, fmt.Errorf("mismatch rich description url suffix, expected suffix is %s but full url is %s", url, richDescription.RenderedURL)
	}

	if res.CreatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected created_at but got zero value")
	}
	if res.UpdatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected updated_at but got zero value")
	}

	// check data in question_hierarchy column of quiz_sets table
	query := fmt.Sprintf(`
		SELECT question_hierarchy FROM quiz_sets
		WHERE lo_id = $1
			AND deleted_at IS NULL`)

	questionHierarchy := &pgtype.JSONBArray{}
	err = s.EurekaDB.QueryRow(ctx, query, &expected.LearningMaterialId).Scan(questionHierarchy)
	if err != nil {
		return ctx, fmt.Errorf("db.QueryRow: %w", err)
	}
	actualQuestionHierarchy := make([]*entities.QuestionHierarchyObj, 0)
	questionHierarchy.AssignTo(&actualQuestionHierarchy)

	// check new question group was added at the end of question_hierarchy list
	if length := len(actualQuestionHierarchy); length != len(stepState.ExistingQuestionHierarchy) {
		return ctx, fmt.Errorf("%s expected %d item in question_hierarchy column of quiz_sets but got %d", stepState.QuestionGroupID, len(stepState.ExistingQuestionHierarchy), length)
	}
	for i := range actualQuestionHierarchy {
		actual := actualQuestionHierarchy[i]
		expected := stepState.ExistingQuestionHierarchy[i]

		if actual.Type != expected.Type {
			return ctx, fmt.Errorf("expected type of item question_hierarchy %d is %s but got %s", i, expected.Type, actual.Type)
		}
		if actual.ID != expected.ID {
			return ctx, fmt.Errorf("expected id of item question_hierarchy %d is %s but got %s", i, expected.ID, actual.ID)
		}
		if len(actual.ChildrenIDs) != len(expected.ChildrenIDs) {
			return ctx, fmt.Errorf("expected children of item question_hierarchy %d is %d but got %d", i, len(expected.ChildrenIDs), len(actual.ChildrenIDs))
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionGroupWasUpdated(ctx context.Context) (context.Context, error) {
	return s.newQuestionGroupWasAddedAtTheEndList(ctx)
}

func (s *Suite) updateAQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.QuestionGroupID) == 0 {
		return ctx, fmt.Errorf("QuestionGroupID dont have yet")
	}

	req := &sspb.UpsertQuestionGroupRequest{
		QuestionGroupId:    stepState.QuestionGroupID,
		LearningMaterialId: stepState.LearningObjectiveID,
		Name:               "name was updated",
		Description:        "description was updated",
		RichDescription: &cpb.RichText{
			Raw:      "raw rich text updated",
			Rendered: "rendered rich text updated",
		},
	}
	stepState.Response, stepState.ResponseErr = s.upsertQuestionGroup(ctx, req)
	stepState.Request = req

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) upsertQuestionGroup(ctx context.Context, req *sspb.UpsertQuestionGroupRequest) (*sspb.UpsertQuestionGroupResponse, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(req.LearningMaterialId) == 0 {
		return nil, fmt.Errorf("lo ID dont have yet")
	}
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	return sspb.NewQuestionServiceClient(s.EurekaConn).
		UpsertQuestionGroup(ctx, req)
}

func (s *Suite) existingQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	token := stepState.Token
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	ctx, err = s.insertANewQuestionGroup(ctx)
	if err != nil {
		return ctx, err
	} else if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	stepState.Token = token
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) existingQuiz(ctx context.Context, template string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	token := stepState.Token
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	req := &epb.UpsertSingleQuizRequest{
		QuizLo: &epb.QuizLO{
			Quiz: &cpb.QuizCore{
				Info: &cpb.ContentBasicInfo{
					SchoolId: constant.ManabieSchool,
					Country:  cpb.Country_COUNTRY_VN,
				},
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Point: wrapperspb.Int32(7),
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Attribute: &cpb.QuizItemAttribute{
							ImgLink:   "img.link",
							AudioLink: "audio.link",
							Configs: []cpb.QuizItemAttributeConfig{
								1,
							},
						},
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Attribute: &cpb.QuizItemAttribute{
							ImgLink:   "img.link",
							AudioLink: "audio.link",
							Configs: []cpb.QuizItemAttributeConfig{
								1,
							},
						},
					},
				},
				Attribute: &cpb.QuizItemAttribute{
					ImgLink:   "img.link",
					AudioLink: "audio.link",
					Configs: []cpb.QuizItemAttributeConfig{
						1,
					},
				},
			},
			LoId: stepState.LearningObjectiveID,
		},
	}

	switch template {
	case "questionGroup":
		req.QuizLo.Quiz.QuestionGroupId = wrapperspb.String(stepState.QuestionGroupID)
	default:
	}

	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	stepState.Request = req
	res, err := epb.NewQuizModifierServiceClient(s.EurekaConn).UpsertSingleQuiz(ctx, req)
	if err != nil {
		return ctx, err
	}
	stepState.Response = res
	stepState.QuizID = req.GetQuizLo().GetQuiz().ExternalId
	stepState.Token = token
	stepState.ExistingQuestionHierarchy.AddQuestionID(req.GetQuizLo().GetQuiz().ExternalId)

	return utils.StepStateToContext(ctx, stepState), nil
}

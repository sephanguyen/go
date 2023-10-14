package eureka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) userUpsertQuiz(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	reqQuiz := &epb.UpsertQuizRequest{
		Quiz: &epb.QuizCore{
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
			SchoolId:   constants.ManabieSchool,
			Country:    cpb.Country_COUNTRY_VN,
			Question: &cpb.RichText{
				Raw:      "Question",
				Rendered: "Question",
			},
			Explanation: &cpb.RichText{
				Raw:      "Explanation",
				Rendered: "Explanation",
			},
			TaggedLos:       []string{"123"},
			DifficultyLevel: 2,
			Options: []*cpb.QuizOption{
				{
					Content: &cpb.RichText{
						Raw:      "Option",
						Rendered: "Option",
					},
					Correctness: true,
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
					},
					Attribute: &cpb.QuizItemAttribute{
						ImgLink:   "https://img.link",
						AudioLink: "https://audio.link",
						Configs: []cpb.QuizItemAttributeConfig{
							cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
						},
					},
				},
			},
			Config: []cpb.QuizConfig{cpb.QuizConfig_QUIZ_CONFIG_OPTIONS_PLAIN_LIST},
		},
		LoId: stepState.LoID,
	}
	switch validity {
	case "missing ExternalId":
		reqQuiz.Quiz.ExternalId = ""
	case "missing Question":
		reqQuiz.Quiz.Question = nil
	case "missing Explanation":
		reqQuiz.Quiz.Explanation = nil
	case "missing Options":
		reqQuiz.Quiz.Options = make([]*cpb.QuizOption, 0)
	}

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertQuiz(ctx, reqQuiz)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) findLoIDOfQuiz(ctx context.Context, quizID pgtype.Text, schoolID pgtype.Int4) (*pgtype.Text, error) {
	query := "SELECT lo_ids FROM quizzes WHERE external_id=$1 AND school_id=$2"
	var loIDs pgtype.TextArray
	err := s.DB.QueryRow(ctx, query, quizID, schoolID).Scan(&loIDs)
	if err != nil {
		return nil, err
	}
	if len(loIDs.Elements) == 0 {
		return nil, fmt.Errorf("cannot find lo ids")
	}

	return &loIDs.Elements[0], nil
}

func (s *suite) userUpsertAQuiz(ctx context.Context, template string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// Save the old token for next steps
	token := stepState.AuthToken

	ctx, err := s.learningObjectiveBelongedToATopic(ctx, "TOPIC_TYPE_EXAM")
	if err != nil {
		return ctx, err
	}
	loID1 := stepState.Request.(string)

	stepState.AuthToken = token
	ctx = s.signedCtx(ctx)
	req := &epb.UpsertQuizRequest{}

	switch template {
	case "valid":
		req = &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
				SchoolId:   constant.ManabieSchool,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{"123", "abc"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
				},
			},
			LoId: loID1,
		}
	case "existed":
		ctx, err := s.userUpsertAQuiz(ctx, "valid")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		prevLoID, err := s.findLoIDOfQuiz(ctx, database.Text(stepState.QuizID), database.Int4(constant.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req = &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: stepState.Request.(*epb.UpsertQuizRequest).Quiz.ExternalId,
				Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
				SchoolId:   constant.ManabieSchool,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{"123", "abc"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
				},
			},
			LoId: prevLoID.String,
		}
	case "emptyTagLO":
		req = &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
				SchoolId:   constant.ManabieSchool,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
					},
				},
			},
			LoId: loID1,
		}
	case "termAndDefinition":
		req = &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_TAD,
				SchoolId:   constant.ManabieSchool,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{"123", "abc"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
				},
			},
			LoId: loID1,
		}
	case "fillInTheBlank":
		req = &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_FIB,
				SchoolId:   constant.ManabieSchool,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{"123", "abc"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
				},
			},
			LoId: loID1,
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertQuiz(ctx, req)
	if stepState.ResponseErr == nil {
		stepState.QuizID = req.Quiz.ExternalId
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) quizCreatedSuccessfullyWithNewVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return ctx, err
	}

	var (
		id       string
		question *cpb.RichText
		options  []*cpb.QuizOption
		kind     cpb.QuizType
	)

	switch req := stepState.Request.(type) {
	case *epb.UpsertQuizRequest:
		question = req.Quiz.Question
		options = req.Quiz.Options
		kind = req.Quiz.Kind
	case *epb.UpsertSingleQuizRequest:
		question = req.QuizLo.Quiz.Question
		options = req.QuizLo.Quiz.Options
		kind = req.QuizLo.Quiz.Kind
	}

	switch resp := stepState.Response.(type) {
	case *epb.UpsertQuizResponse:
		id = resp.Id
	case *epb.UpsertSingleQuizResponse:
		id = resp.Id
	}

	repo := &repositories.QuizRepo{}
	quiz, err := repo.Retrieve(ctx, s.DB, database.Text(id))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	q, _ := quiz.GetQuestion()
	if question.Raw != q.Raw {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s", question.Raw, q.Raw)
	}

	switch req := stepState.Request.(type) {
	case *epb.UpsertSingleQuizRequest:
		point := req.QuizLo.Quiz.Point
		labelType := req.QuizLo.Quiz.LabelType
		if point != nil {
			if quiz.Point.Int != point.Value {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expecting point of quiz %d, got %d", point.Value, quiz.Point.Int)
			}
		} else {
			if quiz.Point.Int != 1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expecting point of quiz %d, got %d", 1, quiz.Point.Int)
			}
		}
		questionTagIds := req.QuizLo.Quiz.QuestionTagIds
		if questionTagIds[0] != stepState.QuestionTagID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting question tag ids to be %s, got %s", stepState.QuestionTagID, questionTagIds[0])
		}

		if labelType.String() != quiz.LabelType.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting label type of quiz %s, got %s", labelType.String(), quiz.LabelType.String)
		}
	}

	opts, _ := quiz.GetOptions()
	if len(options) != len(opts) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %d number of item, got %d", len(options), len(opts))
	}

	for i, opt := range opts {
		if opt.Label != options[i].Label {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz options label %v but got %v", options[i].Label, opt.Label)
		}
		if opt.Key != options[i].Key {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz options key %v but got %v", options[i].Key, opt.Key)
		}
		if opt.Content.Raw != options[i].Content.Raw {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz options content raw %v but got %v", options[i].Content.Raw, opt.Content.Raw)
		}
		if len(opt.Configs) != len(options[i].Configs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz options number of config %v but got %v", len(options[i].Configs), len(opt.Configs))
		}
		if opt.Correctness != options[i].Correctness {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz options correctness %v but got %v", options[i].Correctness, opt.Correctness)
		}
	}

	if kind.String() != quiz.Kind.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("kind %s, got %s", kind.String(), quiz.Kind.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) quizSetAlsoUpdatedWithNewVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		loID string
		id   string
	)

	isUpdateSingleQuiz := false

	switch req := stepState.Request.(type) {
	case *epb.UpsertQuizRequest:
		loID = req.LoId
	case *epb.UpsertSingleQuizRequest:
		isUpdateSingleQuiz = true
		loID = req.QuizLo.LoId
	}

	switch resp := stepState.Response.(type) {
	case *epb.UpsertQuizResponse:
		id = resp.Id
	case *epb.UpsertSingleQuizResponse:
		id = resp.Id
	}

	repo := &repositories.QuizSetRepo{}
	quizsets, err := repo.Search(ctx, s.DB, repositories.QuizSetFilter{
		ObjectiveIDs: database.TextArray([]string{loID}),
		Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
		Limit:        10,
	})
	if err != nil {
		return ctx, err
	}

	found := false
	foundInQuestionHierarchy := false
	qRepo := &repositories.QuizRepo{}
	quiz, err := qRepo.Retrieve(ctx, s.DB, database.Text(id))
	if err != nil {
		return ctx, err
	}

	for _, q := range quizsets {
		for _, externalID := range q.QuizExternalIDs.Elements {
			if quiz.ExternalID.String == externalID.String {
				found = true
				break
			}
		}

		if isUpdateSingleQuiz {
			for _, e := range q.QuestionHierarchy.Elements {
				var questionHierarchyObj *entities.QuestionHierarchyObj
				if err := json.Unmarshal(e.Bytes, &questionHierarchyObj); err != nil {
					return ctx, err
				}

				if quiz.QuestionGroupID.Status == pgtype.Present {
					if quiz.QuestionGroupID.String == questionHierarchyObj.ID {
						for _, extID := range questionHierarchyObj.ChildrenIDs {
							if extID == quiz.ExternalID.String {
								foundInQuestionHierarchy = true
								break
							}
						}
					}
				} else {
					if quiz.ExternalID.String == questionHierarchyObj.ID {
						foundInQuestionHierarchy = true
						break
					}
				}
			}

			if !foundInQuestionHierarchy {
				return ctx, fmt.Errorf("not found new question in question hierarchy")
			}
		}
	}

	if !found {
		return ctx, fmt.Errorf("not found new question in quiz")
	}

	return StepStateToContext(ctx, stepState), nil
}

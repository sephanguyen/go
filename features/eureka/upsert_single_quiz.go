package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) aQuestionTagTypeExistedInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	id := idutil.ULIDNow()
	name := fmt.Sprintf("name+%s", id)
	row := fmt.Sprintf("id,name\n%s,%s", id, name)
	req := &sspb.ImportQuestionTagTypesRequest{
		Payload: []byte(
			row,
		),
	}
	if _, err := sspb.NewQuestionTagTypeClient(s.Conn).ImportQuestionTagTypes(ctx, req); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.ImportQuestionTagTypes, err: %w", err)
	}
	stepState.QuestionTagTypeID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aQuestionTagExistedInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	id := idutil.ULIDNow()
	name := fmt.Sprintf("name+%s", idutil.ULIDNow())
	row := fmt.Sprintf("id,name,question_tag_type_id\n%s,%s,%s", id, name, stepState.QuestionTagTypeID)
	req := &sspb.ImportQuestionTagRequest{
		Payload: []byte(
			row,
		),
	}
	if _, err := sspb.NewQuestionTagClient(s.Conn).ImportQuestionTag(ctx, req); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.ImportQuestionTag, err: %w", err)
	}
	stepState.QuestionTagID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpsertSingleQuiz(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	quiz := cpb.QuizCore{
		Info: &cpb.ContentBasicInfo{
			SchoolId: constants.ManabieSchool,
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
		TaggedLos:       []string{"123"},
		DifficultyLevel: 2,
		Point:           wrapperspb.Int32(4),
		QuestionTagIds:  []string{stepState.QuestionTagID},
		Options: []*cpb.QuizOption{
			{
				Content: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
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
		Attribute: &cpb.QuizItemAttribute{
			ImgLink:   "img.link",
			AudioLink: "audio.link",
			Configs: []cpb.QuizItemAttributeConfig{
				1,
			},
		},
	}
	reqQuiz := &epb.UpsertSingleQuizRequest{
		QuizLo: &epb.QuizLO{
			Quiz: &quiz,
			LoId: stepState.LoID,
		},
	}

	switch validity {
	case "missing ExternalId":
		reqQuiz.QuizLo.Quiz.ExternalId = ""
	case "missing Question":
		reqQuiz.QuizLo.Quiz.Question = nil
	case "missing Explanation":
		reqQuiz.QuizLo.Quiz.Explanation = nil
	case "missing Options":
		reqQuiz.QuizLo.Quiz.Options = make([]*cpb.QuizOption, 0)
	case "missing Attribute":
		reqQuiz.QuizLo.Quiz.Attribute = nil
	}

	stepState.QuestionTagIds = quiz.QuestionTagIds
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertSingleQuiz(ctx, reqQuiz)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpsertAValidSingleQuiz(ctx context.Context, template string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// Save the old token for next steps
	token := stepState.AuthToken

	var loID1 string
	if stepState.LoID == "" {
		ctx, err := s.learningObjectiveBelongedToATopic(ctx, "TOPIC_TYPE_EXAM")
		if err != nil {
			return ctx, err
		}
		loID1 = stepState.Request.(string)
	} else {
		loID1 = stepState.LoID
	}

	stepState.AuthToken = token
	ctx = s.signedCtx(ctx)
	quiz := cpb.QuizCore{
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
		Point:          wrapperspb.Int32(7),
		QuestionTagIds: []string{stepState.QuestionTagID},
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
		DifficultyLevel: 1,
	}
	req := &epb.UpsertSingleQuizRequest{
		QuizLo: &epb.QuizLO{
			Quiz: &quiz,
			LoId: loID1,
		},
	}
	switch template {
	case "existed":
		if ctx, err := s.userUpsertAValidSingleQuiz(ctx, "valid"); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		prevLoID, err := s.findLoIDOfQuiz(ctx, database.Text(stepState.QuizID), database.Int4(constant.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.QuizLo.LoId = prevLoID.String
		req.QuizLo.Quiz.ExternalId = stepState.Request.(*epb.UpsertSingleQuizRequest).QuizLo.Quiz.ExternalId
	case "emptyTagLO":
		req.QuizLo.Quiz.TaggedLos = make([]string, 0)
	case "multipleChoiceQuiz", cpb.QuizType_QUIZ_TYPE_MCQ.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_MCQ
	case "fillInTheBlank", cpb.QuizType_QUIZ_TYPE_FIB.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_FIB
	case "pairOfWordQuiz", cpb.QuizType_QUIZ_TYPE_POW.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_POW
	case "termAndDefinition", cpb.QuizType_QUIZ_TYPE_TAD.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_TAD
	case "manualInputQuiz", cpb.QuizType_QUIZ_TYPE_MIQ.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_MIQ
	case "multiAnswerQuiz", cpb.QuizType_QUIZ_TYPE_MAQ.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_MAQ
	case "orderingQuiz", cpb.QuizType_QUIZ_TYPE_ORD.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_ORD
	case "essayQuiz", cpb.QuizType_QUIZ_TYPE_ESQ.String():
		req.QuizLo.Quiz.Kind = cpb.QuizType_QUIZ_TYPE_ESQ
		req.QuizLo.Quiz.Options = []*cpb.QuizOption{
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
		}
		req.QuizLo.Quiz.AnswerConfig = &cpb.QuizCore_Essay{
			Essay: &cpb.EssayConfig{
				LimitEnabled: true,
				LimitType:    cpb.EssayLimitType_ESSAY_LIMIT_TYPE_CHAR,
				Limit:        5000,
			},
		}
	case "emptyPoint":
		req.QuizLo.Quiz.Point = nil
	case "questionGroup":
		req.QuizLo.Quiz.QuestionGroupId = wrapperspb.String(stepState.QuestionGroupID)
	case "withLabelType":
		req.QuizLo.Quiz.LabelType = cpb.QuizLabelType_QUIZ_LABEL_TYPE_CUSTOM
	}

	stepState.Request = req
	stepState.QuestionTagIds = quiz.QuestionTagIds
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertSingleQuiz(ctx, req)
	if stepState.ResponseErr == nil {
		stepState.QuizID = req.QuizLo.Quiz.ExternalId

		quizEnt := &entities.Quiz{}
		database.AllNullEntity(quizEnt)
		quizEnt.ID = database.Text(stepState.Response.(*epb.UpsertSingleQuizResponse).Id)
		quizEnt.ExternalID = database.Text(quiz.ExternalId)
		quizEnt.Country = database.Text(quiz.Info.Country.String())
		quizEnt.SchoolID = database.Int4(quiz.Info.SchoolId)
		quizEnt.LoIDs = database.TextArray([]string{loID1})
		quizEnt.Kind = database.Text(quiz.Kind.String())
		quizEnt.Question = database.JSONB(&entities.RichText{
			Raw:         quiz.Question.Raw,
			RenderedURL: quiz.Question.Rendered,
		})
		quizEnt.Explanation = database.JSONB(&entities.RichText{
			Raw:         quiz.Explanation.Raw,
			RenderedURL: quiz.Explanation.Rendered,
		})
		quizEnt.Options = database.JSONB(quiz.Options)
		quizEnt.TaggedLOs = database.TextArray(quiz.TaggedLos)
		quizEnt.DifficultLevel = database.Int4(quiz.DifficultyLevel)
		quizEnt.CreatedBy = database.Text(stepState.CurrentUserID)
		quizEnt.ApprovedBy = database.Text(stepState.CurrentUserID)
		quizEnt.Status = database.Text("QUIZ_STATUS_APPROVED")
		quizEnt.Point = database.Int4(quiz.Point.GetValue())
		quizEnt.QuestionGroupID = database.Text(quiz.QuestionGroupId.GetValue())
		quizEnt.LabelType = database.Text(quiz.LabelType.String())

		stepState.Quizzes = append(stepState.Quizzes, quizEnt)
		stepState.ExistingQuestionHierarchy.AddQuestionID(quizEnt.ExternalID.String)
		stepState.GroupedQuizzes = append(stepState.GroupedQuizzes, quizEnt.ExternalID.String)
	}
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) existingQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := stepState.AuthToken
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
			LoId: stepState.LoID,
		},
	}

	ctx = s.signedCtx(ctx)
	stepState.Request = req
	res, err := epb.NewQuizModifierServiceClient(s.Conn).UpsertSingleQuiz(ctx, req)
	if err != nil {
		return ctx, err
	}
	stepState.Response = res
	stepState.QuizID = req.GetQuizLo().GetQuiz().ExternalId
	stepState.AuthToken = token
	stepState.ExistingQuestionHierarchy.AddQuestionID(req.GetQuizLo().GetQuiz().ExternalId)

	return StepStateToContext(ctx, stepState), nil
}

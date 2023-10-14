package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) userUpsertAValidSingleQuizWithRole(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)
	loID1 := stepState.Request.(string)
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
				Point: wrapperspb.Int32(5),
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
			LoId: loID1,
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertSingleQuiz(ctx, req)
	if stepState.ResponseErr == nil {
		stepState.QuizID = req.QuizLo.Quiz.ExternalId
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) learningObjectiveBelongedToTopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.learningObjectiveBelongedToATopic(ctx, "TOPIC_TYPE_EXAM")
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

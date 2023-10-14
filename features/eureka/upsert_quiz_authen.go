package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userUpsertQuizWithRole(ctx context.Context, validity string) (context.Context, error) {
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
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpsertQuiz(ctx, reqQuiz)
	return StepStateToContext(ctx, stepState), nil
}

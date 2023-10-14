package flashcard

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreateAFlashcardContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quizID := idutil.ULIDNow()
	stepState.QuizID = quizID
	externalID := idutil.ULIDNow()
	quizFlashcard := &cpb.QuizCore{
		ExternalId: externalID,
		Kind:       cpb.QuizType_QUIZ_TYPE_POW,
		Info: &cpb.ContentBasicInfo{
			SchoolId: constants.ManabieSchool,
			Country:  cpb.Country_COUNTRY_VN,
		},
		Question: &cpb.RichText{
			Raw:      "{\"blocks\":[{\"key\":\"fhk50\",\"text\":\"D1\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
			Rendered: "rendered " + idutil.ULIDNow(),
		},
		Explanation: &cpb.RichText{
			Raw:      "{\"blocks\":[{\"key\":\"5ksdl\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
			Rendered: "rendered " + idutil.ULIDNow(),
		},
		TaggedLos:       []string{"123", "abc"},
		DifficultyLevel: 2,
		Options: []*cpb.QuizOption{
			{
				Content: &cpb.RichText{Raw: "{\"blocks\":[{\"key\":\"b2ohp\",\"text\":\"T1\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Correctness: false,
				Configs: []cpb.QuizOptionConfig{
					cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
				},
				Attribute: &cpb.QuizItemAttribute{
					ImgLink:   "img.link",
					AudioLink: "audio.link",
					Configs: []cpb.QuizItemAttributeConfig{
						cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
					},
				},
				Label: "label",
				Key:   "key",
			},
		},
		Attribute: &cpb.QuizItemAttribute{
			ImgLink:   "img.link",
			AudioLink: "audio.link",
			Configs: []cpb.QuizItemAttributeConfig{
				cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
			},
		},
		Point: wrapperspb.Int32(10),
	}

	quizFlashcard2 := &cpb.QuizCore{
		ExternalId: externalID,
		Kind:       cpb.QuizType_QUIZ_TYPE_POW,
		Info: &cpb.ContentBasicInfo{
			SchoolId: constants.ManabieSchool,
			Country:  cpb.Country_COUNTRY_VN,
		},
		Question: &cpb.RichText{
			Raw:      "{\"blocks\":[{\"key\":\"2f8iq\",\"text\":\"D2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
			Rendered: "rendered " + idutil.ULIDNow(),
		},
		Explanation: &cpb.RichText{
			Raw:      "{\"blocks\":[{\"key\":\"106qj\",\"text\":\"D3\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
			Rendered: "rendered " + idutil.ULIDNow(),
		},
		TaggedLos:       []string{"123", "abc"},
		DifficultyLevel: 2,
		Options: []*cpb.QuizOption{
			{
				Content: &cpb.RichText{Raw: "{\"blocks\":[{\"key\":\"8am86\",\"text\":\"T2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Correctness: false,
				Configs: []cpb.QuizOptionConfig{
					cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
				},
				Attribute: &cpb.QuizItemAttribute{
					ImgLink:   "img.link",
					AudioLink: "audio.link",
					Configs: []cpb.QuizItemAttributeConfig{
						cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
					},
				},
				Label: "label",
				Key:   "key",
			},
		},
		Attribute: &cpb.QuizItemAttribute{
			ImgLink:   "img.link",
			AudioLink: "audio.link",
			Configs: []cpb.QuizItemAttributeConfig{
				cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
			},
		},
		Point: wrapperspb.Int32(10),
	}

	stepState.QuizFlashcardList = append(stepState.QuizFlashcardList, quizFlashcard)

	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).UpsertFlashcardContent(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.UpsertFlashcardContentRequest{
		Quizzes:     []*cpb.QuizCore{quizFlashcard, quizFlashcard2},
		FlashcardId: stepState.FlashcardID,
		Kind:        cpb.QuizType_QUIZ_TYPE_POW,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateAFlashcardContentWith(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quizID := idutil.ULIDNow()
	stepState.QuizID = quizID
	externalID := idutil.ULIDNow()
	quizFlashcard := &cpb.QuizCore{
		ExternalId: externalID,
		Kind:       cpb.QuizType_QUIZ_TYPE_POW,
		Info: &cpb.ContentBasicInfo{
			SchoolId: constants.ManabieSchool,
			Country:  cpb.Country_COUNTRY_VN,
		},
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
				Content:     &cpb.RichText{Raw: "raw", Rendered: "rendered " + idutil.ULIDNow()},
				Correctness: false,
				Configs: []cpb.QuizOptionConfig{
					cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
				},
				Attribute: &cpb.QuizItemAttribute{
					ImgLink:   "img.link",
					AudioLink: "audio.link",
					Configs: []cpb.QuizItemAttributeConfig{
						cpb.QuizItemAttributeConfig(cpb.QuizItemAttributeConfig_value[arg1]),
					},
				},
				Label: "label",
				Key:   "key",
			},
		},
		Attribute: &cpb.QuizItemAttribute{
			ImgLink:   "img.link",
			AudioLink: "audio.link",
			Configs: []cpb.QuizItemAttributeConfig{
				cpb.QuizItemAttributeConfig(cpb.QuizItemAttributeConfig_value[arg1]),
			},
		},
		Point: wrapperspb.Int32(10),
	}
	stepState.QuizFlashcardList = append(stepState.QuizFlashcardList, quizFlashcard)

	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).UpsertFlashcardContent(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.UpsertFlashcardContentRequest{
		Quizzes:     []*cpb.QuizCore{quizFlashcard},
		FlashcardId: stepState.FlashcardID,
		Kind:        cpb.QuizType_QUIZ_TYPE_POW,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

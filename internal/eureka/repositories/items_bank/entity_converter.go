package repositories

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	entities "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"

	"github.com/google/uuid"
)

func ToLearnosityQuestion(q *entities.ItemsBankQuestion) *learnosity_entity.Question {
	// uuid generator
	refQuestion := uuid.New().String()
	stimulus := ""
	metaData := learnosity_entity.Metadata{}
	explanationHTML := wrapTextWithImageHTML(q.ExplanationText, q.ExplanationImage)

	if q.QuestionType != entities.QuestionTypeFillInTheBlank {
		stimulus = wrapTextWithImageHTML(q.QuestionContentText, q.QuestionContentImage)
	}

	if q.QuestionType == entities.QuestionManualInput {
		metaData.MIQHintExplanation = explanationHTML
	} else {
		metaData.DistractorRationale = explanationHTML
	}
	learnosityQuestionType := getLearnosityQuestionType(q.QuestionType)
	isMath := isQuestionContainsMath(q)
	baseQuestionData := &learnosity_entity.BaseQuestionData{
		Type:             learnosityQuestionType,
		Stimulus:         stimulus,
		Metadata:         metaData,
		FeedbackAttempts: 1,
		InstantFeedback:  true,
		IsMath:           isMath,
	}

	data := getQuestionDataByType(q, baseQuestionData)
	learnosityQuestion := &learnosity_entity.Question{
		Reference: refQuestion,
		Type:      learnosityQuestionType,
		Data:      data,
	}

	return learnosityQuestion
}

func isQuestionContainsMath(q *entities.ItemsBankQuestion) bool {
	if isContainsLatex(q.QuestionContentText) {
		return true
	}
	if isContainsLatex(q.ExplanationText) {
		return true
	}
	for _, v := range q.Options {
		if q.QuestionType != entities.QuestionTypeFillInTheBlank && q.QuestionType != entities.QuestionTypeShortText {
			if isContainsLatex(v.OptionText) {
				return true
			}
		}
	}
	return false
}

func getQuestionDataByType(q *entities.ItemsBankQuestion, baseQuestionData *learnosity_entity.BaseQuestionData) learnosity_entity.QuestionDataInterface {
	switch q.QuestionType {
	case entities.QuestionTypeMultipleChoice, entities.QuestionTypeMultipleAnswers:
		multipleResponses := false
		if q.QuestionType == entities.QuestionTypeMultipleAnswers {
			multipleResponses = true
		}
		learnosityOptions := []learnosity_entity.Options{}
		for i, v := range q.Options {
			learnosityOptions = append(learnosityOptions, learnosity_entity.Options{
				Value: strconv.Itoa(i),
				Label: wrapTextWithImageHTML(v.OptionText, v.OptionImage),
			})
		}

		value := []string{}
		for i, v := range q.CorrectOptions {
			if v {
				value = append(value, strconv.Itoa(i))
			}
		}
		return &learnosity_entity.McqQuestionData{
			BaseQuestionData:  baseQuestionData,
			ShuffleOptions:    true,
			MultipleResponses: multipleResponses,
			Validation: learnosity_entity.McqValidation{
				ScoringType: learnosity_entity.ScoringTypeExactMatch,
				ValidResponse: learnosity_entity.McqResponse{
					Value: value,
					Score: q.Point,
				},
			},
			Options: learnosityOptions,
		}

	case entities.QuestionTypeOrdering:
		optionList := []string{}
		validationValue := []int{}
		for i, v := range q.Options {
			validationValue = append(validationValue, i)
			optionList = append(optionList, wrapTextWithImageHTML(v.OptionText, v.OptionImage))
		}

		return &learnosity_entity.OrdQuestionData{
			BaseQuestionData: baseQuestionData,
			List:             optionList,
			ShuffleOptions:   true,
			Validation: learnosity_entity.OrdValidation{
				ScoringType: learnosity_entity.ScoringTypeExactMatch,
				ValidResponse: learnosity_entity.OrdResponse{
					Value: validationValue,
					Score: q.Point,
				},
			},
		}

	case entities.QuestionTypeFillInTheBlank:
		correctAnswers := strings.Split(q.Options[0].OptionText, ";")
		alternativeResponses := []learnosity_entity.FibResponse{}
		for i := 1; i < len(q.Options); i++ {
			alternativeAnswers := strings.Split(q.Options[i].OptionText, ";")
			alternativeResponses = append(alternativeResponses, learnosity_entity.FibResponse{
				Value: alternativeAnswers,
				Score: q.Point,
			})
		}
		return &learnosity_entity.FibQuestionData{
			BaseQuestionData: baseQuestionData,
			Template:         q.QuestionContentText,
			Validation: learnosity_entity.FibValidation{
				ScoringType: learnosity_entity.ScoringTypeExactMatch,
				ValidResponse: learnosity_entity.FibResponse{
					Value: correctAnswers,
					Score: q.Point,
				},
				AltResponses: alternativeResponses,
			},
		}

	case entities.QuestionTypeShortText:
		altResponses := []learnosity_entity.StqResponse{}
		for i := 1; i < len(q.Options); i++ {
			altResponses = append(altResponses, learnosity_entity.StqResponse{
				Value: q.Options[i].OptionText,
				Score: q.Point,
			})
		}
		return &learnosity_entity.StqQuestionData{
			BaseQuestionData: baseQuestionData,
			Validation: learnosity_entity.StqValidation{
				ScoringType: learnosity_entity.ScoringTypeExactMatch,
				ValidResponse: learnosity_entity.StqResponse{
					Value: q.Options[0].OptionText,
					Score: q.Point,
				},
				AltResponses: altResponses,
			},
		}

	case entities.QuestionManualInput:
		MIQOptions := []learnosity_entity.Options{
			{Value: "0", Label: "True"},
			{Value: "1", Label: "False"},
		}
		MIQValidRes := learnosity_entity.McqResponse{
			Value: []string{"0"},
			Score: q.Point,
		}
		return &learnosity_entity.McqQuestionData{
			BaseQuestionData:  baseQuestionData,
			ShuffleOptions:    false,
			MultipleResponses: false,
			Validation: learnosity_entity.McqValidation{
				ScoringType:   learnosity_entity.ScoringTypeExactMatch,
				ValidResponse: MIQValidRes,
			},
			Options: MIQOptions,
		}

	default:
		return nil
	}
}

func getLearnosityQuestionType(questionType string) string {
	questionTypeMap := map[string]string{
		entities.QuestionTypeMultipleChoice:  learnosity_entity.MultipleChoice,
		entities.QuestionTypeMultipleAnswers: learnosity_entity.MultipleChoice,
		entities.QuestionManualInput:         learnosity_entity.MultipleChoice,
		entities.QuestionTypeOrdering:        learnosity_entity.Ordering,
		entities.QuestionTypeFillInTheBlank:  learnosity_entity.FillInTheBlank,
		entities.QuestionTypeShortText:       learnosity_entity.ShortText,
	}

	return questionTypeMap[questionType]
}

func wrapTextWithImageHTML(textContent string, imageURL string) string {
	if imageURL == "" {
		return textContent
	}
	if textContent == "" {
		return fmt.Sprintf(`<img src="%s" style="max-width:100%%;" />`, imageURL)
	}
	var htmlContent string
	htmlContent += fmt.Sprintf("<p>%s</p>", textContent)

	image := fmt.Sprintf(`<img src="%s" style="max-width:100%%;" />`, imageURL)
	htmlContent += fmt.Sprintf("<p>%s</p>", image)
	return htmlContent
}

func ToLearnosityFeature(i *entities.ItemsBankItem) *learnosity_entity.Feature {
	if strings.TrimSpace(i.ItemName) == "" && strings.TrimSpace(i.ItemDescriptionText) == "" {
		return nil
	}

	isMath := isContainsLatex(i.ItemDescriptionText)

	featureRef := uuid.New().String()
	return learnosity_entity.NewPassageFeature(
		i.ItemName,
		wrapTextWithImageHTML(i.ItemDescriptionText, i.ItemDescriptionImage),
		featureRef,
		isMath,
	)
}

func ToLearnosityItem(i *entities.ItemsBankItem, organizationID string, questionRefs []string, featureReference string) (*learnosity_entity.Item, error) {
	if len(questionRefs) == 0 {
		return nil, fmt.Errorf("question refs must not be empty")
	}
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID must not be empty")
	}
	itemQuestionRefs := []learnosity_entity.Reference{}
	for _, v := range questionRefs {
		itemQuestionRefs = append(itemQuestionRefs, learnosity_entity.Reference{
			Reference: v,
		})
	}
	widgets := []learnosity_entity.Reference{}
	features := []learnosity_entity.Reference{}
	if featureReference != "" {
		featureRef := learnosity_entity.Reference{
			Reference: featureReference,
		}
		features = append(features, featureRef)
		widgets = append(widgets, featureRef)
	}

	for _, v := range itemQuestionRefs {
		widgets = append(widgets,
			learnosity_entity.Reference{
				Reference: v.Reference,
			},
		)
	}

	return &learnosity_entity.Item{
		Reference: i.ItemID,
		Status:    learnosity_entity.ItemStatusPublished,
		Metadata:  nil,
		Features:  features,
		Questions: itemQuestionRefs,
		Definition: learnosity_entity.Definition{
			Widgets: widgets,
		},
		Tags: learnosity_entity.Tags{
			Tenant: []string{organizationID},
		},
	}, nil
}

func isContainsLatex(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	return regexp.MustCompile(`\\\(.*\\\)|\\\[.*\\\]|\$\$.*\$\$`).MatchString(s)
}

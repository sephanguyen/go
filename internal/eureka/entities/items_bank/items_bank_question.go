package services

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	QuestionTypeMultipleChoice  = "MCQ"
	QuestionTypeMultipleAnswers = "MAQ"
	QuestionTypeFillInTheBlank  = "FIB"
	QuestionTypeOrdering        = "ORD"
	QuestionTypeShortText       = "STQ"
	QuestionManualInput         = "MIQ"
)

type ItemsBankOption struct {
	OptionText  string
	OptionImage string
}

type ItemsBankQuestion struct {
	LineNumber           int
	ItemID               string
	QuestionType         string
	Point                int
	QuestionContentText  string
	QuestionContentImage string
	ExplanationText      string
	ExplanationImage     string
	Options              []*ItemsBankOption
	CorrectOptions       []bool
}

func NewItemsBankQuestion(lineNumber int, questionRow *ItemsBankCsvRow) (*ItemsBankQuestion, error) {
	rowPoint := questionRow.Point
	if rowPoint == "" {
		rowPoint = "1"
	}
	point, err := strconv.Atoi(rowPoint)
	if err != nil {
		return nil, err
	}
	question := &ItemsBankQuestion{
		LineNumber:           lineNumber,
		ItemID:               questionRow.ItemID,
		QuestionType:         questionRow.QuestionType,
		Point:                point,
		QuestionContentText:  questionRow.QuestionContentText,
		QuestionContentImage: questionRow.QuestionContentImage,
		ExplanationText:      questionRow.ExplanationText,
		ExplanationImage:     questionRow.ExplanationImage,
		Options:              []*ItemsBankOption{},
		CorrectOptions:       []bool{},
	}

	return question, nil
}

func (q *ItemsBankQuestion) AddOption(optionText string, optionImage string) {
	optionText = strings.TrimSpace(optionText)
	optionImage = strings.TrimSpace(optionImage)
	if optionText == "" && optionImage == "" {
		return
	}
	q.Options = append(q.Options, &ItemsBankOption{
		OptionText:  optionText,
		OptionImage: optionImage,
	})
}
func (q *ItemsBankQuestion) AddCorrectOption(correctOption bool) {
	q.CorrectOptions = append(q.CorrectOptions, correctOption)
}

func (q *ItemsBankQuestion) ValidateNumberOfOptions() error {
	if q.QuestionType == QuestionTypeMultipleChoice ||
		q.QuestionType == QuestionTypeMultipleAnswers ||
		q.QuestionType == QuestionTypeFillInTheBlank {
		if len(q.Options) >= 1 {
			return nil
		}
	}
	if q.QuestionType == QuestionTypeOrdering {
		if len(q.Options) >= 2 {
			return nil
		}
	}
	if q.QuestionType == QuestionTypeShortText {
		if len(q.Options) >= 1 {
			return nil
		}
	}

	if q.QuestionType == QuestionManualInput {
		if len(q.Options) == 0 {
			return nil
		}
	}
	return fmt.Errorf("number of options is invalid")
}

func (q *ItemsBankQuestion) ValidateQuestionType() error {
	s := []string{
		QuestionTypeMultipleChoice,
		QuestionTypeMultipleAnswers,
		QuestionTypeFillInTheBlank,
		QuestionTypeOrdering,
		QuestionTypeShortText,
		QuestionManualInput,
	}
	if !slices.Contains(s, q.QuestionType) {
		return fmt.Errorf("invalid question_type")
	}
	return nil
}

func (q *ItemsBankQuestion) ValidateRequiredFields() error {
	if q.ItemID == "" {
		return fmt.Errorf("missing item_id")
	}
	if q.QuestionType == "" {
		return fmt.Errorf("missing question_type")
	}
	if q.QuestionContentText == "" && q.QuestionContentImage == "" {
		return fmt.Errorf("missing question_content")
	}
	return nil
}

func (q *ItemsBankQuestion) ValidateExplanation() error {
	if q.ExplanationText == "" && q.ExplanationImage == "" {
		return fmt.Errorf("missing explanation_text and explanation_image")
	}
	return nil
}

func (q *ItemsBankQuestion) IsExplanationEmpty() bool {
	return q.ExplanationText == "" && q.ExplanationImage == ""
}

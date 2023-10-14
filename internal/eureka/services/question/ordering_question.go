package question

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

type OrderingQuestion struct {
	*entities.Quiz
	userSubmittedKeys []string
}

func NewOrderingQuestion(quiz *entities.Quiz) *OrderingQuestion {
	return &OrderingQuestion{
		Quiz: quiz,
	}
}

func (o *OrderingQuestion) ResetUserAnswer() {
	o.userSubmittedKeys = []string{}
}

func (o *OrderingQuestion) GetQuizExternalID() string {
	return o.Quiz.ExternalID.String
}

// GetUserAnswerFromSubmitQuizAnswersRequest will mapping user answer ordering from protobuf type to slice of labels
func (o *OrderingQuestion) GetUserAnswerFromSubmitQuizAnswersRequest(submit *epb.SubmitQuizAnswersRequest) (Executor, error) {
	if err := addUserAnswer[*epb.QuizAnswer](o, submit.GetQuizAnswer(), func(quiz *epb.QuizAnswer) error {
		for _, ans := range quiz.Answer {
			answer, isSubmittedKey := ans.Format.(*epb.Answer_SubmittedKey)
			if !isSubmittedKey {
				return fmt.Errorf("your answer is not the ordering type, question %s (external_id), %s (quiz_id)", o.ExternalID.String, o.ID.String)
			}
			o.userSubmittedKeys = append(o.userSubmittedKeys, answer.SubmittedKey)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OrderingQuestion) GetUserAnswerFromCheckQuizCorrectnessRequest(submit []*sspb.CheckQuizCorrectnessRequest) (Executor, error) {
	if err := addUserAnswer[*sspb.CheckQuizCorrectnessRequest](o, submit, func(quiz *sspb.CheckQuizCorrectnessRequest) error {
		for _, ans := range quiz.Answer {
			answer, isSubmittedKey := ans.Format.(*sspb.Answer_SubmittedKey)
			if !isSubmittedKey {
				return fmt.Errorf("your answer is not the ordering type, question %s (external_id), %s (quiz_id)", o.ExternalID.String, o.ID.String)
			}
			o.userSubmittedKeys = append(o.userSubmittedKeys, answer.SubmittedKey)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return o, nil
}

func addUserAnswer[I Input, E Executor](o E, input []I, fn func(I) error) error {
	o.ResetUserAnswer()
	for _, quiz := range input {
		if quiz.GetQuizId() == o.GetQuizExternalID() {
			if err := fn(quiz); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *OrderingQuestion) CheckCorrectness() (*entities.QuizAnswer, error) {
	options, err := o.GetOptions()
	if err != nil {
		return nil, fmt.Errorf("Quiz.GetOptions: %v", err)
	}

	// validate
	if len(options) != len(o.userSubmittedKeys) {
		return nil, fmt.Errorf("number of user's answer must is %d but got %d \nUser's answer: %v", len(options), len(o.userSubmittedKeys), o.userSubmittedKeys)
	}

	// check correctness
	allCorrect := true
	crn := make([]bool, 0, len(o.userSubmittedKeys))
	correctKeys := make([]string, 0, len(options))
	for i, opt := range options {
		correctKey := opt.Key
		correctKeys = append(correctKeys, correctKey)
		if o.userSubmittedKeys[i] != correctKey {
			allCorrect = false
			crn = append(crn, false)
		} else {
			crn = append(crn, true)
		}
	}

	var point uint32
	if allCorrect {
		point = uint32(o.Point.Int)
	}

	return &entities.QuizAnswer{
		QuizID:        o.ExternalID.String,
		QuizType:      o.Kind.String,
		SubmittedKeys: o.userSubmittedKeys, // answer of user, contain option's keys
		CorrectKeys:   correctKeys,         // right answer, contain option's keys
		Correctness:   crn,                 // correctness for each options
		IsAccepted:    allCorrect,          // result for this question
		IsAllCorrect:  allCorrect,          // result for this question
		Point:         point,               // if all order options is correct, this point is point of question, otherwise is 0
		SubmittedAt:   time.Now(),
	}, nil
}

package question

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

type EssayQuestion struct {
	*entities.Quiz
	FilledText []string
}

func NewEssayQuestion(quiz *entities.Quiz) *EssayQuestion {
	return &EssayQuestion{
		Quiz: quiz,
	}
}

func (e *EssayQuestion) ResetUserAnswer() {
	e.FilledText = []string{}
}

func (e *EssayQuestion) GetQuizExternalID() string {
	return e.Quiz.ExternalID.String
}

// GetUserAnswerFromSubmitQuizAnswersRequest will mapping user answer ordering from protobuf type to slice of labels
func (e *EssayQuestion) GetUserAnswerFromSubmitQuizAnswersRequest(submit *epb.SubmitQuizAnswersRequest) (Executor, error) {
	if err := addUserAnswer[*epb.QuizAnswer](e, submit.GetQuizAnswer(), func(quiz *epb.QuizAnswer) error {
		for _, ans := range quiz.Answer {
			answer, isFilledText := ans.Format.(*epb.Answer_FilledText)
			if !isFilledText {
				return fmt.Errorf("your answer is not the essay type, question %s (external_id), %s (quiz_id)", e.ExternalID.String, e.ID.String)
			}
			e.FilledText = append(e.FilledText, answer.FilledText)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *EssayQuestion) GetUserAnswerFromCheckQuizCorrectnessRequest(submit []*sspb.CheckQuizCorrectnessRequest) (Executor, error) {
	if err := addUserAnswer[*sspb.CheckQuizCorrectnessRequest](e, submit, func(quiz *sspb.CheckQuizCorrectnessRequest) error {
		for _, ans := range quiz.Answer {
			answer, isFilledText := ans.Format.(*sspb.Answer_FilledText)
			if !isFilledText {
				return fmt.Errorf("your answer is not the essay type, question %s (external_id), %s (quiz_id)", e.ExternalID.String, e.ID.String)
			}
			e.FilledText = append(e.FilledText, answer.FilledText)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *EssayQuestion) CheckCorrectness() (*entities.QuizAnswer, error) {
	return &entities.QuizAnswer{
		QuizID:      e.ExternalID.String,
		QuizType:    e.Kind.String,
		FilledText:  e.FilledText,
		SubmittedAt: time.Now(),
	}, nil
}

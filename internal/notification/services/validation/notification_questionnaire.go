package validation

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateSubmitQuestionnaire(req *npb.SubmitQuestionnaireRequest,
	questionnaire *entities.Questionnaire,
	questionnaireQuestions entities.QuestionnaireQuestions,
	oldUserAnswerQuestionnaires entities.QuestionnaireUserAnswers) error {
	questionIDs := make([]string, 0, len(questionnaireQuestions))
	for _, question := range questionnaireQuestions {
		questionIDs = append(questionIDs, question.QuestionnaireQuestionID.String)
	}

	countInValidAnswer := 0
	mapQuestionAnswer := make(map[string][]*cpb.Answer, len(questionIDs))
	for _, answer := range req.Answers {
		if slices.Contains(questionIDs, answer.QuestionnaireQuestionId) {
			mapQuestionAnswer[answer.QuestionnaireQuestionId] = append(mapQuestionAnswer[answer.QuestionnaireQuestionId], answer)
		} else {
			countInValidAnswer++
		}
	}

	// Validate invalid answer question
	if countInValidAnswer > 0 {
		return status.Error(codes.InvalidArgument, "you cannot answer the question not in questionnaire")
	}

	for _, q := range questionnaireQuestions {
		// Validate required question
		if q.IsRequired.Bool {
			errRet := status.Error(codes.InvalidArgument, "missing required question, you need to fill all required question")
			answers, ok := mapQuestionAnswer[q.QuestionnaireQuestionID.String]
			if !ok {
				return errRet
			}

			countEmptyAnswer := 0
			for _, answer := range answers {
				if answer.Answer == "" {
					countEmptyAnswer++
				}
			}
			if countEmptyAnswer == len(answers) {
				return errRet
			}
		}

		// Validate answer type
		isTypeHaveOnlyOneAnswer := q.Type.String == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String() || q.Type.String == cpb.QuestionType_QUESTION_TYPE_FREE_TEXT.String()
		if isTypeHaveOnlyOneAnswer && len(mapQuestionAnswer[q.QuestionnaireQuestionID.String]) > 1 {
			return status.Error(codes.InvalidArgument, "you cannot have multiple answer for multiple choices and free text question")
		}

		// Validate choice valid
		isChoiceQuestion := q.Type.String == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String() || q.Type.String == cpb.QuestionType_QUESTION_TYPE_CHECK_BOX.String()
		for _, answer := range mapQuestionAnswer[q.QuestionnaireQuestionID.String] {
			if isChoiceQuestion && !slices.Contains(database.FromTextArray(q.Choices), answer.Answer) {
				return status.Error(codes.InvalidArgument, "your answer doesn't in questionnaire question choices")
			}
		}
	}

	// Validate re-submit
	if !questionnaire.ResubmitAllowed.Bool && len(oldUserAnswerQuestionnaires) > 0 {
		return status.Error(codes.InvalidArgument, "resubmit not allowed, you cannot re-submit this questionnaire")
	}
	return nil
}

func ValidateQuestionnairePb(qn *cpb.Questionnaire, noti *cpb.Notification) error {
	if qn.ExpirationDate == nil {
		return fmt.Errorf("you cannot set expiration date of questionnaire is empty")
	}

	now := time.Now().Truncate(time.Minute).Add(time.Minute)
	qnEndDate := qn.ExpirationDate.AsTime()

	if qnEndDate.Before(now) {
		return fmt.Errorf("you cannot set expiration date of questionnaire at a time in the past")
	}

	if noti.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED {
		scheduledAt := noti.ScheduledAt.AsTime()
		if qnEndDate.Before(scheduledAt) {
			return fmt.Errorf("you cannot set the expiration date of a questionnaire at a time before notification's schedule time")
		}
	}

	if qn.Questions == nil || len(qn.Questions) == 0 {
		return fmt.Errorf("you cannot upsert a questionnaire without any questions")
	}

	for _, question := range qn.Questions {
		if question.Title == "" {
			return fmt.Errorf("you cannot set question's title is empty")
		}

		isTypeMustHaveChoices := question.Type == cpb.QuestionType_QUESTION_TYPE_CHECK_BOX || question.Type == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE
		isChoicesQuestionNil := question.Choices == nil || len(question.Choices) < 2
		if isTypeMustHaveChoices && isChoicesQuestionNil {
			return fmt.Errorf("question with type is multiple choice or check box must have least two choices")
		}

		if isTypeMustHaveChoices {
			for _, choice := range question.Choices {
				if choice == "" {
					return fmt.Errorf("you cannot set question's choice is empty")
				}
			}
		}
	}

	return nil
}

func ValidateExportQuestionnaireAnswersCSV(supportedLangs []string, req *npb.GetQuestionnaireAnswersCSVRequest) (*time.Location, error) {
	if !slices.Contains(supportedLangs, req.Language) {
		return nil, fmt.Errorf("your language: %s is not supported", req.Language)
	}

	clientLocation, err := time.LoadLocation(req.Timezone)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client timezone: "+err.Error())
	}

	return clientLocation, nil
}

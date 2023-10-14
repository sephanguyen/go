package mappers

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	exportentities "github.com/manabie-com/backend/internal/notification/export_entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func QuestionnaireUserAnswersToExportedCSVResponders(responders []*repositories.QuestionnaireCSVResponder, questionnaireUserAnswers entities.QuestionnaireUserAnswers, questionnaireQuestions entities.QuestionnaireQuestions) []*exportentities.QuestionnaireCSVResponder {
	resp := make([]*exportentities.QuestionnaireCSVResponder, 0)

	mapUserIDAndUserAnswers := make(map[string]entities.QuestionnaireUserAnswers, 0)
	for _, qNUserAnswer := range questionnaireUserAnswers {
		mapUserIDAndUserAnswers[qNUserAnswer.UserID.String] = append(mapUserIDAndUserAnswers[qNUserAnswer.UserID.String], qNUserAnswer)
	}

	for _, responder := range responders {
		userAnswer := &exportentities.QuestionnaireCSVResponder{
			ResponderName:      responder.Name.String,
			UserID:             responder.UserID.String,
			TargetID:           responder.TargetID.String,
			StudentID:          responder.StudentID.String,
			StudentExternalID:  responder.StudentExternalID.String,
			TargetName:         responder.TargetName.String,
			UserNotificationID: responder.UserNotificationID.String,
			IsParent:           responder.IsParent.Bool,
			IsIndividual:       responder.IsIndividual.Bool,
			LocationNames:      database.FromTextArray(responder.LocationNames),
		}

		// Fill submitted_at field
		submittedAt := database.FromTimestamptz(responder.SubmittedAt)
		if submittedAt != nil {
			userAnswer.SubmittedAt = *submittedAt
		}

		userAnswersEnt, ok := mapUserIDAndUserAnswers[responder.UserID.String]
		if ok {
			userAnswerSubmitedAt := time.Now()
			answers := make([]*exportentities.QuestionnaireAnswer, 0)
			for _, userAnswerEnt := range userAnswersEnt {
				if userAnswerEnt.TargetID.String == userAnswer.TargetID {
					answers = append(answers, &exportentities.QuestionnaireAnswer{
						QuestionnaireQuestionID: userAnswerEnt.QuestionnaireQuestionID.String,
						Answer:                  userAnswerEnt.Answer.String,
					})
					userAnswerSubmitedAt = userAnswerEnt.SubmittedAt.Time
				}
			}

			// userAnswer.SubmittedAt is set above, but for some old submitted, it will be null -> need to get from user answer and assign for it.
			if userAnswer.SubmittedAt.IsZero() && responder.SubmissionStatus.String == cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String() {
				userAnswer.SubmittedAt = userAnswerSubmitedAt
			}
			userAnswer.QuestionnaireAnswers = answers
		}

		resp = append(resp, userAnswer)
	}
	return resp
}

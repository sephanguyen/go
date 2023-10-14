package mappers

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	exportentities "github.com/manabie-com/backend/internal/notification/export_entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"

	"github.com/stretchr/testify/assert"
)

func Test_QuestionnaireUserAnswersToExportedCSVResponders(t *testing.T) {
	t.Parallel()

	timeSubmitted := time.Now()
	responders := []*repositories.QuestionnaireCSVResponder{
		{
			UserNotificationID: database.Text("user-notification-1"),
			UserID:             database.Text("student-1"),
			TargetID:           database.Text("student-1"),
			StudentID:          database.Text("student-1"),
			Name:               database.Text("student-name-1"),
			TargetName:         database.Text("student-name-1"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			IsParent:           database.Bool(false),
			IsIndividual:       database.Bool(false),
			StudentExternalID:  database.Text("student-external-id-1"),
			LocationNames:      database.TextArray([]string{"location-1"}),
		},
		{
			UserNotificationID: database.Text("user-notification-2"),
			UserID:             database.Text("parent-1"),
			TargetID:           database.Text("student-1"),
			StudentID:          database.Text("student-1"),
			Name:               database.Text("parent-name-1"),
			TargetName:         database.Text("student-name-1"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			StudentExternalID:  database.Text("student-external-id-1"),
			IsParent:           database.Bool(true),
			IsIndividual:       database.Bool(false),
			LocationNames:      database.TextArray([]string{"location-2"}),
		},
		{
			UserNotificationID: database.Text("user-notification-2"),
			UserID:             database.Text("parent-1"),
			TargetID:           database.Text("parent-1"),
			StudentID:          database.Text(""),
			Name:               database.Text("parent-name-1"),
			TargetName:         database.Text("parent-name-1"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			StudentExternalID:  database.Text(""),
			IsParent:           database.Bool(true),
			IsIndividual:       database.Bool(true),
			LocationNames:      database.TextArray([]string{"location-2"}),
		},
	}

	questionnaireUserAnswers := entities.QuestionnaireUserAnswers{}
	questionnaireQuestions := utils.GenQNQuestions(idutil.ULIDNow())

	checkResponders := func(t *testing.T, responders []*repositories.QuestionnaireCSVResponder, userAnswersCSV []*exportentities.QuestionnaireCSVResponder) {
		for idx, userAnswerCSV := range userAnswersCSV {
			responder := responders[idx]

			assert.Equal(t, userAnswerCSV.UserNotificationID, responder.UserNotificationID.String)
			assert.Equal(t, userAnswerCSV.UserID, responder.UserID.String)
			assert.Equal(t, userAnswerCSV.TargetID, responder.TargetID.String)
			assert.Equal(t, userAnswerCSV.StudentID, responder.StudentID.String)
			assert.Equal(t, userAnswerCSV.ResponderName, responder.Name.String)
			assert.Equal(t, userAnswerCSV.TargetID, responder.TargetID.String)
			assert.Equal(t, userAnswerCSV.TargetName, responder.TargetName.String)
			assert.Equal(t, userAnswerCSV.SubmittedAt.UTC(), responder.SubmittedAt.Time.UTC())
			assert.Equal(t, userAnswerCSV.IsParent, responder.IsParent.Bool)
			assert.Equal(t, userAnswerCSV.IsIndividual, responder.IsIndividual.Bool)
			assert.Equal(t, userAnswerCSV.LocationNames, database.FromTextArray(responder.LocationNames))
		}
	}

	checkQuesionnaireUserAnswers := func(t *testing.T, questionnaireUserAnswers entities.QuestionnaireUserAnswers, userAnswersCSV []*exportentities.QuestionnaireCSVResponder, questionnaireQuestions entities.QuestionnaireQuestions) {
		// Collect all responders answers to check valid data
		answers := []*exportentities.QuestionnaireAnswer{}
		for _, userAnswerCSV := range userAnswersCSV {
			answers = append(answers, userAnswerCSV.QuestionnaireAnswers...)
		}

		countCorrect := 0
		for _, questionnaireUserAnswer := range questionnaireUserAnswers {
			for _, answer := range answers {
				if answer.Answer == questionnaireUserAnswer.Answer.String &&
					answer.QuestionnaireQuestionID == questionnaireUserAnswer.QuestionnaireQuestionID.String {
					countCorrect++
				}
			}
		}
		assert.Equal(t, len(answers), countCorrect)
	}

	t.Run("happy case", func(t *testing.T) {
		for _, responder := range responders {
			for _, questionnaireQuestion := range questionnaireQuestions {
				answer := utils.GenQNUserAnswer(responder.UserID.String, responder.TargetID.String, questionnaireQuestion.QuestionnaireQuestionID.String)
				questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
			}
		}

		userAnswersCSV := QuestionnaireUserAnswersToExportedCSVResponders(responders, questionnaireUserAnswers, questionnaireQuestions)
		checkResponders(t, responders, userAnswersCSV)
		checkQuesionnaireUserAnswers(t, questionnaireUserAnswers, userAnswersCSV, questionnaireQuestions)
	})
}

package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	exportentities "github.com/manabie-com/backend/internal/notification/export_entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/notification/services/validation"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	QuestionnaireAnswerCSVTitleLanguagesMapping = map[string][]string{
		"vi": {"", "Timestamp", "Location", "Responder Name", "Associated Student", "Student ID", "External Student ID"},
		"en": {"", "Timestamp", "Location", "Responder Name", "Associated Student", "Student ID", "External Student ID"},
		"ja": {"", "回答日時", "所属", "回答者", "生徒名", "生徒ID", "外部生徒ID"},
	}

	QuestionCSVTitleLanguagesMapping = map[string]string{
		"vi": "Question",
		"en": "Question",
		"ja": "質問",
	}
)

func (svc *NotificationReaderService) GetQuestionnaireAnswersCSV(ctx context.Context, req *npb.GetQuestionnaireAnswersCSVRequest) (*npb.GetQuestionnaireAnswersCSVResponse, error) {
	supportedLanguages := make([]string, 0, len(QuestionCSVTitleLanguagesMapping))
	for k := range QuestionCSVTitleLanguagesMapping {
		supportedLanguages = append(supportedLanguages, k)
	}

	clientLocation, err := validation.ValidateExportQuestionnaireAnswersCSV(supportedLanguages, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validate get questionnaire answers csv: "+err.Error())
	}

	// Get questions by questionnaire_id
	questionnaireQuestions, err := svc.QuestionnaireRepo.FindQuestionsByQnID(ctx, svc.DB, req.QuestionnaireId)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding questions: "+err.Error())
	}
	if len(questionnaireQuestions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No questions found with questionnaireId: "+req.QuestionnaireId)
	}

	questionnaireQuestionIDs := make([]string, 0)
	for _, q := range questionnaireQuestions {
		questionnaireQuestionIDs = append(questionnaireQuestionIDs, q.QuestionnaireQuestionID.String)
	}

	questionnaireCSVResponders, err := svc.QuestionnaireRepo.FindQuestionnaireCSVResponders(ctx, svc.DB, req.QuestionnaireId)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding questionnaire responders: "+err.Error())
	}

	responderIDs := make([]string, 0)
	targetIDs := make([]string, 0)
	for _, questionnaireResponder := range questionnaireCSVResponders {
		responderIDs = append(responderIDs, questionnaireResponder.UserID.String)
		targetIDs = append(targetIDs, questionnaireResponder.TargetID.String)
	}

	// Get questionnaire user answers using found responders and target answered questionnaire
	findUserAnswersFilter := repositories.NewFindUserAnswersFilter()
	findUserAnswersFilter.QuestionnaireQuestionIDs = database.TextArray(questionnaireQuestionIDs)
	findUserAnswersFilter.UserIDs = database.TextArray(responderIDs)
	findUserAnswersFilter.TargetIDs = database.TextArray(targetIDs)
	userAnswers, err := svc.QuestionnaireRepo.FindUserAnswers(ctx, svc.DB, &findUserAnswersFilter)

	exportedCSVResponders := mappers.QuestionnaireUserAnswersToExportedCSVResponders(questionnaireCSVResponders, userAnswers, questionnaireQuestions)
	csvArrayData := svc.exportCSVRespondersToCSVArrayData(exportedCSVResponders, questionnaireQuestions, clientLocation)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "error when convert to csv: "+err.Error())
	}

	title := QuestionnaireAnswerCSVTitleLanguagesMapping[req.Language]
	for idx := range questionnaireQuestions {
		title = append(title, fmt.Sprintf("%s %d", QuestionCSVTitleLanguagesMapping[req.Language], idx+1))
	}
	csvData := append([][]string{title}, csvArrayData...)

	return &npb.GetQuestionnaireAnswersCSVResponse{
		Data: exporter.ToCSV(csvData),
	}, nil
}

func (svc *NotificationReaderService) exportCSVRespondersToCSVArrayData(questionnaireCSVResponders []*exportentities.QuestionnaireCSVResponder, questionnaireQuestions entities.QuestionnaireQuestions, clientLocation *time.Location) [][]string {
	csvArrayData := make([][]string, 0)

	for idx, questionnaireCSVResponder := range questionnaireCSVResponders {
		submittedAt := ""
		if !questionnaireCSVResponder.SubmittedAt.IsZero() {
			submittedAt = questionnaireCSVResponder.SubmittedAt.UTC().In(clientLocation).Format(consts.DateTimeCSVFormat)
		}
		associatedStudent := ""
		if questionnaireCSVResponder.IsParent && !questionnaireCSVResponder.IsIndividual {
			associatedStudent = questionnaireCSVResponder.TargetName
		}
		row := []string{
			fmt.Sprint(idx + 1),
			submittedAt,
			strings.Join(questionnaireCSVResponder.LocationNames, ", "),
			questionnaireCSVResponder.ResponderName,
			associatedStudent,
			questionnaireCSVResponder.StudentID,
			questionnaireCSVResponder.StudentExternalID,
		}

		mapQuestionIDAndAnswer := make(map[string][]*exportentities.QuestionnaireAnswer)
		for _, answer := range questionnaireCSVResponder.QuestionnaireAnswers {
			mapQuestionIDAndAnswer[answer.QuestionnaireQuestionID] = append(mapQuestionIDAndAnswer[answer.QuestionnaireQuestionID], answer)
		}
		for _, question := range questionnaireQuestions {
			answersStr := make([]string, 0)
			answers, ok := mapQuestionIDAndAnswer[question.QuestionnaireQuestionID.String]
			if !ok || len(answers) == 0 {
				// empty answer
				row = append(row, "")
				continue
			}

			switch question.Type.String {
			case cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String(),
				cpb.QuestionType_QUESTION_TYPE_CHECK_BOX.String():
				for _, answer := range answers {
					idxAnswer := slices.Index(database.FromTextArray(question.Choices), answer.Answer)
					answerChar, err := utils.ConvertNumberToUppercaseChar(idxAnswer)
					if err == nil {
						answersStr = append(answersStr, answerChar)
					}
				}
				sort.Slice(answersStr, func(i, j int) bool {
					return answersStr[i] <= answersStr[j]
				})
			case cpb.QuestionType_QUESTION_TYPE_FREE_TEXT.String():
				answersStr = append(answersStr, answers[0].Answer)
			}

			row = append(row, strings.Join(answersStr, ", "))
		}

		csvArrayData = append(csvArrayData, row)
	}

	return csvArrayData
}

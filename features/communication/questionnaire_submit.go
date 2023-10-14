package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QuestionaireSubmitSuite struct {
	*common.NotificationSuite
	parentID                   string
	parentToken                string
	createdQN                  *cpb.Questionnaire
	mapStudentUserNotification map[string]string
	answers                    []*cpb.Answer
	errSubmitAPI               error
}

func (c *SuiteConstructor) InitQuestionnaireSubmit(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &QuestionaireSubmitSuite{
		NotificationSuite:          dep.notiCommonSuite,
		mapStudentUserNotification: make(map[string]string),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff create a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                  s.CurrentStaffCreateQuestionnaire,
		`^current staff creates "([^"]*)" students with the same parent$`:                                                           s.CreatesNumberOfStudentsWithSameParentsInfo,
		`^current staff upsert notification and send to parent$`:                                                                    s.currentStaffUpsertNotificationAndSendToParent,
		`^parent does not see answers in previous step calling RetrieveNotificationDetail with notification for student (\d+)$`:     s.parentDoesNotSeeAnswersForStudent,
		`^parent login to Learner App$`:                           s.parentLoginToLearnerApp,
		`^update user device token to an "([^"]*)" device token$`: s.UpdateDeviceTokenForLeanerUser,
		`^parent see answers in previous step calling RetrieveNotificationDetail with notification for student (\d+)$`: s.parentSeeAnswersForStudent,
		`^parent see answers in previous step calling RetrieveNotificationDetail with notification for themself`:       s.parentSeeAnswersForThemself,
		`^parent submit answers list for questions in questionnaire for student (\d+) with "([^"]*)" answers$`:         s.parentSubmitAnswersListForQuestionsInQuestionnaireForStudent,
		`^parent submit answers list for questions in questionnaire for themself with "([^"]*)" answers$`:              s.parentSubmitAnswersListForQuestionsInQuestionnaireForThemself,
		`^current staff creates "([^"]*)" students with "([^"]*)" parent info$`:                                        s.CreatesNumberOfStudentsWithParentsInfo,
		`^change expiration_date of questionnaire in db to yesterday$`:                                                 s.changeExpirationDateOfQuestionnaireInDBToYesterday,
		`^parent receive "([^"]*)" error$`:                                                 s.parentReceiveError,
		`^parent re-submit answers list for questions in questionnaire for student (\d+)$`: s.parentReSubmitQuestionsInQuestionnaireForStudent,
		`^update answers list suitable with "([^"]*)" error$`:                              s.updateAnswersListSuitableWithError,
		`^current staff set target for notification of student (\d+) to parent$`:           s.currentStaffSetTargetForNotificationOfStudentToParent,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *QuestionaireSubmitSuite) parentLoginToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.parentID = commonState.Students[0].Parents[0].ID
	var err error
	s.parentToken, err = s.GenerateExchangeTokenCtx(ctx, s.parentID, commonState.Students[0].Parents[0].Group)
	if err != nil {
		return ctx, nil
	}
	return ctx, nil
}

func (s *QuestionaireSubmitSuite) currentStaffUpsertNotificationAndSendToParent(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	receiverIDs := []string{}
	for _, student := range commonState.Students {
		receiverIDs = append(receiverIDs, student.ID)
	}
	opts := &common.NotificationWithOpts{
		UserGroups:       "parent",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "none",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
		ReceiverIds:      receiverIDs,
	}

	ctx, err := s.CurrentStaffUpsertNotificationWithOpts(ctx, opts)
	if err != nil {
		return ctx, err
	}

	ctx, err = s.CurrentStaffSendNotification(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *QuestionaireSubmitSuite) parentSubmitAnswersListForQuestionsInQuestionnaireForThemself(ctx context.Context, ansType string) (context.Context, error) {
	return s.parentSubmitAnswersListForQuestionsInQuestionnaireForStudent(ctx, -1, ansType)
}

func (s *QuestionaireSubmitSuite) parentSubmitAnswersListForQuestionsInQuestionnaireForStudent(ctx context.Context, studentIndex int, ansType string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	targetID := ""
	if studentIndex < 0 {
		targetID = s.parentID
	} else {
		targetID = commonState.Students[studentIndex].ID
	}

	// parent read noti
	notiDetailRequest := &npb.RetrieveNotificationDetailRequest{
		NotificationId: commonState.Notification.NotificationId,
		TargetId:       targetID,
	}

	notiDetail, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
		common.ContextWithToken(ctx, s.parentToken),
		notiDetailRequest,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed RetrieveNotificationDetail for notiID %s, targetID %s: %v", notiDetailRequest.NotificationId, notiDetailRequest.TargetId, err)
	}

	// parent answer questionnaire
	questionsRes := notiDetail.UserQuestionnaire.Questionnaire.Questions
	for _, qRes := range questionsRes {
		for _, q := range commonState.Questionnaire.Questions {
			if qRes.OrderIndex == q.OrderIndex {
				q.QuestionnaireQuestionId = qRes.QuestionnaireQuestionId
			}
		}
	}

	answers := make([]*cpb.Answer, 0, len(questionsRes))
	for _, ques := range questionsRes {
		switch ansType {
		case "full":
			answers = append(answers, s.MakeAnAnswer(ques))
		case "missing":
			if !ques.Required {
				answers = append(answers, s.MakeAnAnswer(ques))
			}
		case "you cannot answer the question not in questionnaire":
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: idutil.ULIDNow(),
				Answer:                  "Wrong Answer",
			})
		case "you cannot have multiple answer for multiple choices and free text question":
			for _, ques := range questionsRes {
				if ques.Type == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE {
					answers = append(answers, &cpb.Answer{
						QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
						Answer:                  ques.Choices[2],
					})
				}
				if ques.Type == cpb.QuestionType_QUESTION_TYPE_FREE_TEXT {
					answers = append(answers, &cpb.Answer{
						QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
						Answer:                  "Wrong Answer",
					})
				}
			}
		case "your answer doesn't in questionnaire question choices":
			if ques.Type == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE || ques.Type == cpb.QuestionType_QUESTION_TYPE_CHECK_BOX {
				answers = append(answers, &cpb.Answer{
					QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
					Answer:                  idutil.ULIDNow(),
				})
			}
		}
	}

	// parent submit answers
	req := &npb.SubmitQuestionnaireRequest{
		UserInfoNotificationId: notiDetail.UserNotification.UserNotificationId,
		QuestionnaireId:        commonState.Questionnaire.QuestionnaireId,
		Answers:                answers,
	}
	s.answers = answers

	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SubmitQuestionnaire(
		common.ContextWithToken(ctx, s.parentToken), req)
	if err != nil {
		s.errSubmitAPI = err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionaireSubmitSuite) changeExpirationDateOfQuestionnaireInDBToYesterday(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	// update expiration_date of questionnaire in the past
	yesterday := time.Now().AddDate(0, 0, -1)
	query := `
		UPDATE questionnaires
		SET expiration_date = $1
		WHERE questionnaire_id = $2
	`

	formatTimeLayout := "2006-01-02 15:04:05.000000"
	_, err := s.BobDBConn.Exec(ctx, query, yesterday.Format(formatTimeLayout), commonState.Questionnaire.QuestionnaireId)
	if err != nil {
		return ctx, fmt.Errorf("s.BobDBTrace.Exec: %v", err)
	}
	return ctx, nil
}

func (s *QuestionaireSubmitSuite) parentReceiveError(ctx context.Context, errStr string) (context.Context, error) {
	if errStr != "none" {
		if s.errSubmitAPI == nil {
			return ctx, fmt.Errorf("expected an error when submit: %s, but didn't get any error", errStr)
		}
	} else {
		if s.errSubmitAPI == nil {
			return ctx, nil
		}
	}

	errSubmitCheck := status.Error(codes.InvalidArgument, errStr)
	if errSubmitCheck.Error() != s.errSubmitAPI.Error() {
		return ctx, fmt.Errorf("expected an error when submit: %s, but got: %s", errSubmitCheck.Error(), s.errSubmitAPI.Error())
	}

	return ctx, nil
}

func (s *QuestionaireSubmitSuite) parentReSubmitQuestionsInQuestionnaireForStudent(ctx context.Context, studentIndex int) (context.Context, error) {
	return s.parentSubmitAnswersListForQuestionsInQuestionnaireForStudent(ctx, studentIndex, "full")
}

func (s *QuestionaireSubmitSuite) parentDoesNotSeeAnswersForStudent(ctx context.Context, studentIndex int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	notiInfo, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
		common.ContextWithToken(ctx, s.parentToken), &npb.RetrieveNotificationDetailRequest{
			NotificationId: commonState.Notification.NotificationId,
			TargetId:       commonState.Students[studentIndex].ID,
		})
	if err != nil {
		return ctx, err
	}

	if notiInfo.UserQuestionnaire.IsSubmitted {
		return ctx, fmt.Errorf("submitted questionnaire, expected: questionnaire isn't submitted")
	}

	return ctx, nil
}

func (s *QuestionaireSubmitSuite) parentSeeAnswersForThemself(ctx context.Context) (context.Context, error) {
	return s.parentSeeAnswersForStudent(ctx, -1)
}

func (s *QuestionaireSubmitSuite) parentSeeAnswersForStudent(ctx context.Context, studentIndex int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	targetID := ""
	if studentIndex < 0 {
		targetID = s.parentID
	} else {
		targetID = commonState.Students[studentIndex].ID
	}

	notiInfo, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
		common.ContextWithToken(ctx, s.parentToken),
		&npb.RetrieveNotificationDetailRequest{
			NotificationId: commonState.Notification.NotificationId,
			TargetId:       targetID,
		})

	if err != nil {
		return ctx, err
	}

	if !notiInfo.UserQuestionnaire.IsSubmitted {
		return ctx, fmt.Errorf("doesn't submit questionnaire, expected: questionnaire is submitted")
	}

	if len(s.answers) != len(notiInfo.UserQuestionnaire.Answers) {
		return ctx, fmt.Errorf("missing some answers when submitted")
	}

	return ctx, nil
}

func (s *QuestionaireSubmitSuite) updateAnswersListSuitableWithError(ctx context.Context, errStr string) (context.Context, error) {
	switch errStr {
	case "you cannot answer the question not in questionnaire":
		s.answers = append(s.answers, &cpb.Answer{
			QuestionnaireQuestionId: idutil.ULIDNow(),
			Answer:                  "Wrong Answer",
		})
	case "you cannot have multiple answer for multiple choices and free text question":
		for _, ques := range s.createdQN.Questions {
			if ques.Type == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE {
				s.answers = append(s.answers, &cpb.Answer{
					QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
					Answer:                  ques.Choices[2],
				})
			}
			if ques.Type == cpb.QuestionType_QUESTION_TYPE_FREE_TEXT {
				s.answers = append(s.answers, &cpb.Answer{
					QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
					Answer:                  "Wrong Answer",
				})
			}
		}
	case "your answer doesn't in questionnaire question choices":
		for _, ques := range s.createdQN.Questions {
			if ques.Type == cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE || ques.Type == cpb.QuestionType_QUESTION_TYPE_CHECK_BOX {
				for _, ans := range s.answers {
					if ans.QuestionnaireQuestionId == ques.QuestionnaireQuestionId {
						ans.Answer = idutil.ULIDNow()
					}
				}
			}
		}
	}
	return ctx, nil
}

func (s *QuestionaireSubmitSuite) currentStaffSetTargetForNotificationOfStudentToParent(ctx context.Context, studentIndex int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	// parent read noti
	notiDetailRequest := &npb.RetrieveNotificationDetailRequest{
		NotificationId: commonState.Notification.NotificationId,
		TargetId:       commonState.Students[studentIndex].ID,
	}

	notiDetail, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
		common.ContextWithToken(ctx, s.parentToken),
		notiDetailRequest,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed RetrieveNotificationDetail for notiID %s, targetID %s: %v", notiDetailRequest.NotificationId, notiDetailRequest.TargetId, err)
	}

	stmt := `
		UPDATE users_info_notifications SET student_id = NULL, parent_id = $2
		WHERE user_notification_id = $1
		AND deleted_at IS NULL;
	`
	_, err = s.BobDBConn.Exec(ctx, stmt, notiDetail.UserNotification.UserNotificationId, s.parentID)
	if err != nil {
		return ctx, nil
	}

	return ctx, nil
}

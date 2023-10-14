package communication

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuestionnaireDownloadCSV struct {
	*common.NotificationSuite

	parentID                     string
	parentName                   string
	createdQN                    *cpb.Questionnaire
	answeredNotiIDs              []string
	mapUserIDAndSubmittedAnswers map[string][]*cpb.Answer
	studentNamesCreated          []string
	mapStudentName               map[string]string
	questionnaireQuestions       []*cpb.Question

	mapGenericUserIDAndUserGroup map[string]string
	mapUserIDAndName             map[string]string
}

func (c *SuiteConstructor) InitQuestionnaireDownloadCsv(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &QuestionnaireDownloadCSV{
		NotificationSuite:            dep.notiCommonSuite,
		mapUserIDAndSubmittedAnswers: make(map[string][]*cpb.Answer),
		mapStudentName:               make(map[string]string, 0),
		mapGenericUserIDAndUserGroup: make(map[string]string, 0),
		mapUserIDAndName:             make(map[string]string, 0),
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with the same parent$`:                                                            s.CreatesNumberOfStudentsWithSameParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                                       s.aQuestionnaireWithResubmitAllowedQuestionsRespectively,
		`^current staff send a notification with attached questionnaire to "([^"]*)" and individual "([^"]*)"$`:                     s.currentStaffSendANotificationWithAttachedQuestionnaireToAndIndividual,
		`^notificationmgmt services must send notification to user$`:                                                                s.NotificationMgmtMustSendNotificationToUser,
		`^"([^"]*)" see "([^"]*)" unanswered questionnaire in notification bell with correct detail$`:                               s.seeUnansweredQuestionnaireInNotificationBellWithCorrectDetail,
		`^parent answer questionnaire for "([^"]*)"$`:                                                                               s.parentAnswerQuestionnaire,
		`^students answer questionnaire for themselves$`:                                                                            s.studentsAnswerQuestionnaireForThemselves,
		`^"([^"]*)" answer questionnaire for themself$`:                                                                             s.answerQuestionnaireForThemself,
		`^parent answer questionnaire for themself$`:                                                                                s.parentAnswerQuestionnaireForThemself,
		`^current staff download questionnaire answers csv file successfully with "([^"]*)" rows and "([^"]*)" submissions$`:        s.currentStaffDownloadQuestionnaireAnswersCsvFile,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *QuestionnaireDownloadCSV) parentAnswerQuestionnaire(ctx context.Context, answerFor string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.parentID = commonState.Students[0].Parents[0].ID
	s.parentName = commonState.Students[0].Parents[0].Name
	if s.parentID == "" {
		return ctx, fmt.Errorf("parentID is empty")
	}
	token, err := s.GenerateExchangeTokenCtx(ctx, s.parentID, constant.UserGroupParent)
	ctxWithToken := s.ContextWithToken(ctx, token)
	if err != nil {
		return ctx, err
	}

	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctxWithToken, &npb.RetrieveNotificationsRequest{
		Paging: &cpb.Paging{Limit: 100},
	})
	if err != nil {
		return ctx, err
	}

	studentIDMap := make(map[string]string)
	for index, student := range commonState.Students {
		studentIDMap[strconv.Itoa(index+1)] = student.ID
	}

	stuNotiIDMap := map[string]string{}
	for _, item := range res.Items {
		stuNotiIDMap[item.TargetId] = item.UserNotification.UserNotificationId
		s.createdQN.QuestionnaireId = item.QuestionnaireId
	}

	studentIdxes := strings.Split(answerFor, ",")
	userNotiIDs := make([]string, 0, len(studentIdxes))
	for _, answeredStu := range studentIdxes {
		stuID := studentIDMap[answeredStu]
		userNotiIDs = append(userNotiIDs, stuNotiIDMap[stuID])
	}

	err = s.userAnswerQuesionnaire(ctx, s.parentID, commonState.Students[0].ID, userNotiIDs)

	return ctx, err
}

func (s *QuestionnaireDownloadCSV) seeUnansweredQuestionnaireInNotificationBellWithCorrectDetail(ctx context.Context, person, num string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	numQn, _ := strconv.Atoi(num)

	mapUserIDAndUserGroup, err := s.getMapUserIDAndUserGroup(commonState, person)
	if err != nil {
		return ctx, err
	}
	for userID, userGroup := range mapUserIDAndUserGroup {
		token, err := s.GenerateExchangeTokenCtx(ctx, userID, userGroup)
		ctxWithToken := s.ContextWithToken(ctx, token)
		if err != nil {
			return ctx, err
		}
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctxWithToken, &npb.RetrieveNotificationsRequest{
			Paging: &cpb.Paging{Limit: 100},
		})
		if err != nil {
			return ctx, err
		}
		if len(res.Items) != numQn {
			return ctx, fmt.Errorf("want %d notification with questionnaire, has %d items in the bell, check user %s", numQn, len(res.Items), userID)
		}
		for _, item := range res.Items {
			notiInfo, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(ctxWithToken, &npb.RetrieveNotificationDetailRequest{
				NotificationId: commonState.Notification.NotificationId,
				TargetId:       item.TargetId,
			})
			if err != nil {
				return ctx, err
			}

			if protoEqual(notiInfo.UserQuestionnaire.Questionnaire, s.createdQN) {
				return ctx, fmt.Errorf("diff between create questionnaire and viewed qn %s, check user %s", protoDiff(notiInfo.UserQuestionnaire.Questionnaire, s.createdQN), userID)
			}

			if notiInfo.UserQuestionnaire.IsSubmitted || len(notiInfo.UserQuestionnaire.Answers) > 0 {
				return ctx, fmt.Errorf("questionnaire is already answered somehow, check user %s", userID)
			}
		}
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireDownloadCSV) currentStaffSendANotificationWithAttachedQuestionnaireToAndIndividual(ctx context.Context, receiver, individualReceivers string) (context.Context, error) {
	var commonState *common.StepState
	opts := &common.NotificationWithOpts{
		CourseFilter:     "all",
		GradeFilter:      "all",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "none",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
	}
	switch receiver {
	case "none":
		opts.CourseFilter = "none"
		opts.GradeFilter = "none"
		opts.UserGroups = "student, parent"
	case "student":
		opts.UserGroups = "student"
	case "parent":
		opts.UserGroups = "parent"
	default:
		opts.UserGroups = "student, parent"
	}

	genericUserIDs := []string{}
	switch individualReceivers {
	case "student individual":
		// create 1 student and use as generic receiver id
		ctx, err = s.CreatesNumberOfStudents(ctx, "1")
		if err != nil {
			return ctx, fmt.Errorf("failed CreatesNumberOfStudents: %v", err)
		}
		commonState = common.StepStateFromContext(ctx)
		for idx, student := range commonState.Students {
			if idx >= len(commonState.Students)-1 {
				genericUserIDs = append(genericUserIDs, student.ID)
				s.mapGenericUserIDAndUserGroup[student.ID] = cpb.UserGroup_USER_GROUP_STUDENT.String()
			}
		}
	case "parent individual":
		// create 1 student with 1 parent but only use parent
		ctx, err = s.CreatesNumberOfStudentsWithParentsInfo(ctx, "1", "1")
		if err != nil {
			return ctx, fmt.Errorf("failed CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
		commonState = common.StepStateFromContext(ctx)
		for idx, student := range commonState.Students {
			if idx >= len(commonState.Students)-1 {
				for _, parent := range student.Parents {
					genericUserIDs = append(genericUserIDs, parent.ID)
					s.mapGenericUserIDAndUserGroup[parent.ID] = cpb.UserGroup_USER_GROUP_PARENT.String()
				}
			}
		}
	default: // all individual
		// create 1 student with 1 parent and use as generic receiver ids
		ctx, err = s.CreatesNumberOfStudentsWithParentsInfo(ctx, "1", "1")
		if err != nil {
			return ctx, fmt.Errorf("failed CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
		commonState = common.StepStateFromContext(ctx)
		for idx, student := range commonState.Students {
			if idx >= len(commonState.Students)-1 {
				genericUserIDs = append(genericUserIDs, student.ID)
				s.mapGenericUserIDAndUserGroup[student.ID] = cpb.UserGroup_USER_GROUP_STUDENT.String()
				for _, parent := range student.Parents {
					genericUserIDs = append(genericUserIDs, parent.ID)
					s.mapGenericUserIDAndUserGroup[parent.ID] = cpb.UserGroup_USER_GROUP_PARENT.String()
				}
			}
		}
	}
	opts.GenericReceiverIds = genericUserIDs

	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return ctx, fmt.Errorf("failed GetNotificationWithOptions: %v", err)
	}

	req := &npb.UpsertNotificationRequest{
		Notification:  commonState.Notification,
		Questionnaire: s.createdQN,
	}

	res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), req)
	if err != nil {
		return ctx, fmt.Errorf("UpsertNotification %s", err)
	}

	commonState.Notification.NotificationId = res.NotificationId
	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.SendNotificationRequest{
		NotificationId: res.NotificationId,
	})
	if err != nil {
		return ctx, fmt.Errorf("SendNotification %s", err)
	}

	for _, student := range commonState.Students {
		s.mapUserIDAndName[student.ID] = student.Name
		for _, parent := range student.Parents {
			s.mapUserIDAndName[parent.ID] = parent.Name
		}
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireDownloadCSV) aQuestionnaireWithResubmitAllowedQuestionsRespectively(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	questions := parseQuestionFromString(questionStr)
	qn := &cpb.Questionnaire{
		ResubmitAllowed: common.StrToBool(resubmit),
		Questions:       questions,
		ExpirationDate:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	s.createdQN = qn
	return ctx, nil
}

func (s *QuestionnaireDownloadCSV) studentsAnswerQuestionnaireForThemselves(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, student := range commonState.Students {
		s.mapStudentName[student.ID] = student.Name
		s.studentNamesCreated = append(s.studentNamesCreated, student.Name)
		tok, err := s.GenerateExchangeTokenCtx(ctx, student.ID, cpb.UserGroup_USER_GROUP_STUDENT.String())
		if err != nil {
			return ctx, err
		}
		ctxWithToken := s.ContextWithToken(ctx, tok)
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctxWithToken, &npb.RetrieveNotificationsRequest{
			Paging: &cpb.Paging{Limit: 100},
		})
		if err != nil {
			return ctx, err
		}

		if len(res.Items) != 1 {
			return ctx, fmt.Errorf("expected %d user notificaion for student %s, got %d", 1, student.ID, len(res.Items))
		}

		s.createdQN.QuestionnaireId = res.Items[0].QuestionnaireId
		userNotiID := res.Items[0].UserNotification.UserNotificationId

		err = s.userAnswerQuesionnaire(ctx, student.ID, student.ID, []string{userNotiID})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *QuestionnaireDownloadCSV) fillEmptyAnswersForQuestion(answers []*cpb.Answer) []*cpb.Answer {
	mapQuestionIDAndAnswer := make(map[string][]*cpb.Answer)

	for _, answer := range answers {
		mapQuestionIDAndAnswer[answer.QuestionnaireQuestionId] = append(mapQuestionIDAndAnswer[answer.QuestionnaireQuestionId], answer)
	}

	for _, question := range s.questionnaireQuestions {
		_, ok := mapQuestionIDAndAnswer[question.QuestionnaireQuestionId]
		if !ok {
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: question.QuestionnaireQuestionId,
				Answer:                  "",
			})
		}
	}

	return answers
}

func (s *QuestionnaireDownloadCSV) userAnswerQuesionnaire(ctx context.Context, userID, targetID string, userNotiIDs []string) error {
	commonState := common.StepStateFromContext(ctx)
	token, err := s.GenerateExchangeTokenCtx(ctx, userID, constant.UserGroupStudent)
	ctxWithToken := s.ContextWithToken(ctx, token)

	if err != nil {
		return err
	}

	notiInfo, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(ctxWithToken, &npb.RetrieveNotificationDetailRequest{
		NotificationId: commonState.Notification.NotificationId,
		TargetId:       targetID,
	})
	if err != nil {
		return err
	}

	questions := notiInfo.UserQuestionnaire.Questionnaire.Questions

	s.questionnaireQuestions = questions
	answers := makeAnswersListForOnlyRequiredQuestion(questions)

	s.mapUserIDAndSubmittedAnswers[userID] = answers

	for _, userNoti := range userNotiIDs {
		submitReq := &npb.SubmitQuestionnaireRequest{
			QuestionnaireId:        s.createdQN.QuestionnaireId,
			Answers:                answers,
			UserInfoNotificationId: userNoti,
		}
		_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SubmitQuestionnaire(ctxWithToken, submitReq)
		if err != nil {
			return err
		}
	}
	s.answeredNotiIDs = append(s.answeredNotiIDs, userNotiIDs...)
	return nil
}

func (s *QuestionnaireDownloadCSV) checkQuestionnaireData(ctx context.Context, requestQuestinnaire *cpb.Questionnaire, userViewQuestionnaire *cpb.Questionnaire) (context.Context, error) {
	countQuestionnaireQuestion := 0

	for _, reqQuestion := range requestQuestinnaire.Questions {
		for _, userViewQuestion := range userViewQuestionnaire.Questions {
			if ok, _ := protoEqualWithoutOrder(reqQuestion, userViewQuestion, []string{"QuestionnaireQuestionId"}); ok {
				countQuestionnaireQuestion++
			}
		}
	}

	if countQuestionnaireQuestion != len(requestQuestinnaire.Questions) {
		return ctx, fmt.Errorf("created questionnaire question doesn't match with user view quesionnaire question")
	}

	if requestQuestinnaire.ResubmitAllowed != userViewQuestionnaire.ResubmitAllowed {
		return ctx, fmt.Errorf("created questionnaire resubmit allowed doesn't match with user view quesionnaire resubmit allowed, want %v, got %v", requestQuestinnaire.ResubmitAllowed, userViewQuestionnaire.ResubmitAllowed)
	}

	// because round when convert => we need truncate
	reqExpirationDate := requestQuestinnaire.ExpirationDate.AsTime().Truncate(time.Microsecond)
	userViewExpirationDate := userViewQuestionnaire.ExpirationDate.AsTime().Truncate(time.Microsecond)
	if reqExpirationDate != userViewExpirationDate {
		return ctx, fmt.Errorf("created questionnaire expiration date doesn't match with user view quesionnaire expiration date, want %v, got %v", reqExpirationDate, userViewExpirationDate)
	}

	return ctx, nil
}

func (s *QuestionnaireDownloadCSV) parentAnswerQuestionnaireForThemself(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	s.parentID = commonState.Students[0].Parents[0].ID
	s.parentName = commonState.Students[0].Parents[0].Name
	parentToken, err := s.GenerateExchangeTokenCtx(ctx, s.parentID, commonState.Students[0].Parents[0].Group)
	if err != nil {
		return ctx, fmt.Errorf("failed GenerateExchangeTokenCtx: %v", err)
	}

	notiList, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(common.ContextWithToken(ctx, parentToken), &npb.RetrieveNotificationsRequest{
		Paging: &cpb.Paging{Limit: 100},
	})
	if err != nil {
		return ctx, fmt.Errorf("failed RetrieveNotificationDetail for notiID %s, targetID %s: %v", commonState.Notification.NotificationId, s.parentID, err)
	}

	userNotiIDs := make([]string, 0)
	for _, item := range notiList.Items {
		s.createdQN.QuestionnaireId = item.QuestionnaireId
		userNotiIDs = append(userNotiIDs, item.UserNotification.UserNotificationId)
	}

	err = s.userAnswerQuesionnaire(ctx, s.parentID, s.parentID, userNotiIDs)
	if err != nil {
		return ctx, fmt.Errorf("failed userAnswerQuesionnaire for user notiID %s, targetID %s: %v", userNotiIDs, s.parentID, err)
	}

	return ctx, nil
}

func (s *QuestionnaireDownloadCSV) answerQuestionnaireForThemself(ctx context.Context, person string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	mapUserIDAndUserGroup, err := s.getMapUserIDAndUserGroup(commonState, person)
	if err != nil {
		return ctx, err
	}

	for userID, userGroup := range mapUserIDAndUserGroup {
		token, err := s.GenerateExchangeTokenCtx(ctx, userID, userGroup)
		if err != nil {
			return ctx, fmt.Errorf("s.GenerateExchangeTokenCtx: %v", err)
		}

		ctxWithToken := s.ContextWithToken(ctx, token)

		notiList, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(
			ctxWithToken, &npb.RetrieveNotificationsRequest{
				Paging: &cpb.Paging{Limit: 100},
			})
		if err != nil {
			return ctx, fmt.Errorf("failed RetrieveNotificationDetail for notiID %s, targetID %s: %v", commonState.Notification.NotificationId, userID, err)
		}

		userNotiIDs := make([]string, 0)
		for _, item := range notiList.Items {
			s.createdQN.QuestionnaireId = item.QuestionnaireId
			userNotiIDs = append(userNotiIDs, item.UserNotification.UserNotificationId)
		}

		err = s.userAnswerQuesionnaire(ctx, userID, userID, userNotiIDs)
		if err != nil {
			return ctx, fmt.Errorf("failed userAnswerQuesionnaire for user notiID %s, targetID %s: %v", userNotiIDs, userID, err)
		}
	}

	return ctx, nil
}

func (s *QuestionnaireDownloadCSV) getMapUserIDAndUserGroup(state *common.StepState, person string) (map[string]string, error) {
	mapUserIDAndUserGroup := make(map[string]string)
	switch person {
	case "parent":
		for _, student := range state.Students {
			for _, parent := range student.Parents {
				// avoid using generic user id in TargetGroup
				if _, existed := s.mapGenericUserIDAndUserGroup[parent.ID]; !existed {
					mapUserIDAndUserGroup[parent.ID] = parent.Group
				}
			}
		}
	case "student":
		for _, student := range state.Students {
			// avoid using generic user id in TargetGroup
			if _, existed := s.mapGenericUserIDAndUserGroup[student.ID]; !existed {
				mapUserIDAndUserGroup[student.ID] = student.Group
			}
		}
	case "student individual":
		for userID, userGroup := range s.mapGenericUserIDAndUserGroup {
			if userGroup == cpb.UserGroup_USER_GROUP_STUDENT.String() {
				mapUserIDAndUserGroup[userID] = userGroup
			}
		}
	case "parent individual":
		for userID, userGroup := range s.mapGenericUserIDAndUserGroup {
			if userGroup == cpb.UserGroup_USER_GROUP_PARENT.String() {
				mapUserIDAndUserGroup[userID] = userGroup
			}
		}
	case "all individual":
		if len(s.mapGenericUserIDAndUserGroup) == 0 {
			return nil, fmt.Errorf("mapGenericUserIDAndUserGroup is empty")
		}
		for userID, userGroup := range s.mapGenericUserIDAndUserGroup {
			mapUserIDAndUserGroup[userID] = userGroup
		}
	}

	return mapUserIDAndUserGroup, nil
}

func (s *QuestionnaireDownloadCSV) currentStaffDownloadQuestionnaireAnswersCsvFile(ctx context.Context, numRowsStr, numSubmissionsStr string) (context.Context, error) {
	numRows, _ := strconv.Atoi(numRowsStr)
	numSubmissions, _ := strconv.Atoi(numSubmissionsStr)

	commonState := common.StepStateFromContext(ctx)
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).GetQuestionnaireAnswersCSV(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.GetQuestionnaireAnswersCSVRequest{
		QuestionnaireId: s.createdQN.QuestionnaireId,
		Timezone:        "Asia/Ho_Chi_Minh",
		Language:        "en",
	})

	if commonState.ResponseErr != nil {
		return ctx, fmt.Errorf("error download questionnaire answers CSV file: %v", err)
	}
	if len(res.GetData()) == 0 {
		return ctx, fmt.Errorf("expected questionnaire answers CSV file not empty")
	}

	csvFile := res.GetData()
	sc := scanner.NewCSVScanner(bytes.NewReader(csvFile))
	timestamps := []string{}
	responderNames := []string{}
	for sc.Scan() {
		timestamp := sc.Text("Timestamp")
		responderName := sc.Text("Responder Name")

		if timestamp != "" {
			timestamps = append(timestamps, timestamp)
		}

		responderNames = append(responderNames, responderName)
	}

	if len(timestamps) != numSubmissions {
		return ctx, fmt.Errorf("expected %d submission, got %d submission", numSubmissions, len(timestamps))
	}

	if len(responderNames) != numRows {
		return ctx, fmt.Errorf("expected %d rows, got %d rows", numSubmissions, len(timestamps))
	}

	return common.StepStateToContext(ctx, commonState), nil
}

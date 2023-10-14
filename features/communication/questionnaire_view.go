package communication

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	npbv2 "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v2"

	"github.com/cucumber/godog"
	"github.com/r3labs/diff/v3"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuestionaireViewSuite struct {
	*common.NotificationSuite

	parentID                     string
	parentName                   string
	createdQN                    *cpb.Questionnaire
	answeredNotiIDs              []string
	mapUserIDAndSubmittedAnswers map[string][]*cpb.Answer
	studentNamesCreated          []string
	resultAnswersPagination      *npb.GetAnswersByFilterResponse
	mapStudentName               map[string]string
	questionnaireQuestions       []*cpb.Question

	mapGenericUserIDAndUserGroup map[string]string
	mapUserIDAndName             map[string]string
}

var (
	studentNamesV2 = map[string][]string{}
	err            error
)

func (c *SuiteConstructor) InitQuestionnaireView(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &QuestionaireViewSuite{
		NotificationSuite:            dep.notiCommonSuite,
		mapUserIDAndSubmittedAnswers: make(map[string][]*cpb.Answer),
		mapStudentName:               make(map[string]string, 0),
		mapGenericUserIDAndUserGroup: make(map[string]string, 0),
		mapUserIDAndName:             make(map[string]string, 0),
	}
	stepsMapping := map[string]interface{}{
		`^school admin creates "([^"]*)" students with the same parent$`:                                                            s.CreatesNumberOfStudentsWithSameParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^returns "([^"]*)" status code$`:                                                                                           s.CheckReturnStatusCode,
		`^notificationmgmt services must send notification to user$`:                                                                s.NotificationMgmtMustSendNotificationToUser,
		`^a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                                       s.aQuestionnaireWithResubmitAllowedQuestionsRespectively,
		`^"([^"]*)" see "([^"]*)" unanswered questionnaire in notification bell with correct detail$`:                               s.seeUnansweredQuestionnaireInNotificationBellWithCorrectDetail,
		`^current staff send a notification with attached questionnaire to "([^"]*)" and individual "([^"]*)"$`:                     s.currentStaffSendANotificationWithAttachedQuestionnaireToAndIndividual,
		`^parent answer questionnaire for "([^"]*)"$`:                                                                               s.parentAnswerQuestionnaire,
		`^parent answer questionnaire for "([^"]*)" with empty answer$`:                                                             s.parentAnswerQuestionnaireWithEmptyAnswer,
		`^"([^"]*)" see (\d+) notifications in notification bell with correct detail and answer status$`:                            s.seeQuestionnaireInNotificationBellWithCorrectDetailAndAnswerStatus,
		`^current staff see "([^"]*)" answers in questionnaire answers list with search_text is a full name of target "([^"]*)" and total answers is "([^"]*)" and fully answer for question$`: s.currentStaffSeeAnswersInQuestionnaireAnswersListWithSearchTarget,
		`^students answer questionnaire for themselves$`:                                   s.studentsAnswerQuestionnaireForThemselves,
		`^current staff get answers list with limit is "([^"]*)" and offset is "([^"]*)"$`: s.currentStaffGetAnswersListWithOffsetAndLimit,
		`^current staff see "([^"]*)" answers in questionnaire answers list and total items is "([^"]*)" and previous offset is "([^"]*)" and next offset is "([^"]*)"$`: s.checkPaginationData,
		`^current staff set target for user notification of student to parent$`:                                                                                          s.currentStaffSetTargetForUserNotificationOfStudentToParent,
		`^parent answer questionnaire for themself$`:                                                                                                                     s.parentAnswerQuestionnaireForThemself,
		`^"([^"]*)" answer questionnaire for themself$`:                                                                                                                  s.answerQuestionnaireForThemself,
		`^"([^"]*)" see "([^"]*)" unanswered questionnaire in notification bell with correct detail using new api RetrieveNotificationDetail$`:                           s.seeQuestionnaireUsingNewAPI,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *QuestionaireViewSuite) seeQuestionnaireInNotificationBellWithCorrectDetailAndAnswerStatus(ctx context.Context, person string, qnNum int) (context.Context, error) {
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

		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctxWithToken, &npb.RetrieveNotificationsRequest{Paging: &cpb.Paging{Limit: 100}})
		if err != nil {
			return ctx, err
		}
		if len(res.Items) != qnNum {
			return ctx, fmt.Errorf("want %d notification, has %d", qnNum, len(res.Items))
		}
		for _, item := range res.Items {
			info, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(ctxWithToken, &npb.RetrieveNotificationDetailRequest{
				NotificationId: item.UserNotification.NotificationId,
				TargetId:       item.TargetId,
			})
			if err != nil {
				return ctx, err
			}

			expectIsSubmitted := slices.Contains(s.answeredNotiIDs, info.UserNotification.UserNotificationId)
			if expectIsSubmitted != info.UserQuestionnaire.IsSubmitted {
				return ctx, fmt.Errorf("%s want user_info_noti has submit status %v, actual status is %v", person, expectIsSubmitted, info.UserQuestionnaire.IsSubmitted)
			}

			if expectIsSubmitted {
				submittedAnswers := s.mapUserIDAndSubmittedAnswers[userID]
				if ok, diff := protoEqualWithoutOrder(submittedAnswers, info.UserQuestionnaire.Answers, nil); !ok {
					return ctx, fmt.Errorf("%s see answers info not match answer submitted, diff: %s", person, diff)
				}
			} else if len(info.UserQuestionnaire.Answers) > 0 {
				return ctx, fmt.Errorf("unanswer user questionnaire still has answer %v in response", info.UserQuestionnaire.Answers)
			}

			ctx, err = s.checkQuestionnaireData(ctx, s.createdQN, info.UserQuestionnaire.Questionnaire)
			if err != nil {
				return ctx, fmt.Errorf("checkQuestionnaireData: %v", err)
			}
		}
	}
	return ctx, nil
}

func (s *QuestionaireViewSuite) parentAnswerQuestionnaire(ctx context.Context, answerFor string) (context.Context, error) {
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

func (s *QuestionaireViewSuite) parentAnswerQuestionnaireWithEmptyAnswer(ctx context.Context, answerFor string) (context.Context, error) {
	return s.parentAnswerQuestionnaire(ctx, answerFor)
}

func (s *QuestionaireViewSuite) seeUnansweredQuestionnaireInNotificationBellWithCorrectDetail(ctx context.Context, person, num string) (context.Context, error) {
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

func (s *QuestionaireViewSuite) seeQuestionnaireUsingNewAPI(ctx context.Context, person, num string) (context.Context, error) {
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
			notiInfo, err := npbv2.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
				ctxWithToken,
				&npbv2.RetrieveNotificationDetailRequest{
					UserNotificationId: item.UserNotification.UserNotificationId,
				},
			)
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

func (s *QuestionaireViewSuite) currentStaffSendANotificationWithAttachedQuestionnaireToAndIndividual(ctx context.Context, receiver, individualReceivers string) (context.Context, error) {
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

func (s *QuestionaireViewSuite) aQuestionnaireWithResubmitAllowedQuestionsRespectively(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	questions := parseQuestionFromString(questionStr)
	qn := &cpb.Questionnaire{
		ResubmitAllowed: common.StrToBool(resubmit),
		Questions:       questions,
		ExpirationDate:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	s.createdQN = qn
	return ctx, nil
}

func (s *QuestionaireViewSuite) currentStaffSeeAnswersInQuestionnaireAnswersListWithSearchTarget(ctx context.Context, numAnswer int, searchTarget string, totalCount int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	searchText := ""
	switch searchTarget {
	case "all":
		searchText = ""
	case "student":
		studentNames := []string{}
		for _, student := range commonState.Students {
			studentNames = append(studentNames, student.Name)
		}
		// searchText = common.GetRandomKeywordFromStrings(studentNames)
		searchText = studentNames[util.RandRangeIn(0, len(studentNames))]
	case "parent":
		searchText = s.parentName
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).GetAnswersByFilter(
		s.ContextWithToken(ctx, commonState.CurrentStaff.Token),
		&npb.GetAnswersByFilterRequest{
			QuestionnaireId: s.createdQN.QuestionnaireId,
			Keyword:         searchText,
			Paging: &cpb.Paging{
				Limit: math.MaxUint32,
			},
		})
	if err != nil {
		return ctx, fmt.Errorf("GetAnswersByFilter %s", err)
	}

	if len(res.UserAnswers) != numAnswer {
		return ctx, fmt.Errorf("expected %d answers from users, got %d", numAnswer, len(res.UserAnswers))
	}

	if int(res.TotalItems) != totalCount {
		return ctx, fmt.Errorf("expected %d total answers from users, got %d", totalCount, int(res.TotalItems))
	}

	for _, userAnswer := range res.UserAnswers {
		var userIDAnswered string
		// Responder is not from IndividualTarget
		if _, exists := s.mapGenericUserIDAndUserGroup[userAnswer.UserId]; !exists {
			if userAnswer.IsParent {
				// students have same parent
				userIDAnswered = commonState.Students[0].Parents[0].ID
			} else {
				for _, student := range commonState.Students {
					if userAnswer.UserId == student.ID {
						userIDAnswered = student.ID
					}
				}
			}
		} else {
			userIDAnswered = userAnswer.UserId
		}
		answersSubmitted := s.fillEmptyAnswersForQuestion(s.mapUserIDAndSubmittedAnswers[userIDAnswered])
		ok, diff := protoEqualWithoutOrder(answersSubmitted, userAnswer.Answers, nil, diff.AllowTypeMismatch(true), diff.DisableStructValues())
		if !ok {
			fmt.Printf("\nmapUserIDAndSubmittedAnswers: %v\n", s.mapUserIDAndSubmittedAnswers[userIDAnswered])
			fmt.Printf("\nanswersSubmitted: %v\n", answersSubmitted)
			fmt.Printf("\nuserAnswer.Answers: %v\n", userAnswer.Answers)
			return ctx, fmt.Errorf("school admin see student's answers info not match answer submitted of user_id: %s, diff: %s", userAnswer.UserId, diff)
		}

		if userAnswer.ResponderName != s.mapUserIDAndName[userAnswer.UserId] {
			return ctx, fmt.Errorf("school admin see responder name doesn't match, expected %s, got %s", s.mapUserIDAndName[userAnswer.UserId], userAnswer.ResponderName)
		}

		if userAnswer.UserNotificationId == "" {
			return ctx, errors.New("failed to get user_notification_id for user answer")
		}

		// Check order answer for each responder
		mapQuestionIDAndOrderIndex := make(map[string]int64)
		for _, question := range s.questionnaireQuestions {
			mapQuestionIDAndOrderIndex[question.QuestionnaireQuestionId] = question.OrderIndex
		}

		for i := 0; i < len(userAnswer.Answers)-1; i++ {
			if mapQuestionIDAndOrderIndex[userAnswer.Answers[i].QuestionnaireQuestionId] > mapQuestionIDAndOrderIndex[userAnswer.Answers[i+1].QuestionnaireQuestionId] {
				return ctx, errors.New("school admin see user answer doesn't order by question order index asc")
			}
		}
	}

	return ctx, nil
}

func (s *QuestionaireViewSuite) studentsAnswerQuestionnaireForThemselves(ctx context.Context) (context.Context, error) {
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

func (s *QuestionaireViewSuite) currentStaffGetAnswersListWithOffsetAndLimit(ctx context.Context, limit int, offset int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	commonState.Response, commonState.ResponseErr = npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).GetAnswersByFilter(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.GetAnswersByFilterRequest{
		QuestionnaireId: s.createdQN.QuestionnaireId,
		Keyword:         "",
		Paging: &cpb.Paging{
			Limit:  uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(offset)},
		},
	})
	if commonState.ResponseErr == nil {
		s.resultAnswersPagination = commonState.Response.(*npb.GetAnswersByFilterResponse)
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionaireViewSuite) checkPaginationData(ctx context.Context, numberOfResult int, totalCount int, previousOffset int, nextOffset int) (context.Context, error) {
	if int(s.resultAnswersPagination.TotalItems) != totalCount {
		return ctx, fmt.Errorf("expected %d of total answers from users, got %d", totalCount, int(s.resultAnswersPagination.TotalItems))
	}

	if len(s.resultAnswersPagination.UserAnswers) != numberOfResult {
		return ctx, fmt.Errorf("expected %d answers from users, got %d", numberOfResult, len(s.resultAnswersPagination.UserAnswers))
	}

	if int(s.resultAnswersPagination.NextPage.GetOffsetInteger()) != nextOffset {
		return ctx, fmt.Errorf("expected next offset is %d, got %d", nextOffset, int(s.resultAnswersPagination.NextPage.GetOffsetInteger()))
	}

	if int(s.resultAnswersPagination.PreviousPage.GetOffsetInteger()) != previousOffset {
		return ctx, fmt.Errorf("expected previous offset is %d, got %d", previousOffset, int(s.resultAnswersPagination.PreviousPage.GetOffsetInteger()))
	}

	return ctx, nil
}

func (s *QuestionaireViewSuite) fillEmptyAnswersForQuestion(answers []*cpb.Answer) []*cpb.Answer {
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

func (s *QuestionaireViewSuite) userAnswerQuesionnaire(ctx context.Context, userID, targetID string, userNotiIDs []string) error {
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

func (s *QuestionaireViewSuite) checkQuestionnaireData(ctx context.Context, requestQuestinnaire *cpb.Questionnaire, userViewQuestionnaire *cpb.Questionnaire) (context.Context, error) {
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

func (s *QuestionaireViewSuite) currentStaffSetTargetForUserNotificationOfStudentToParent(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.parentID = commonState.Students[0].Parents[0].ID
	parentToken, err := s.GenerateExchangeTokenCtx(ctx, s.parentID, commonState.Students[0].Parents[0].Group)
	if err != nil {
		return ctx, fmt.Errorf("failed GenerateExchangeTokenCtx: %v", err)
	}

	notiDetail, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(common.ContextWithToken(ctx, parentToken), &npb.RetrieveNotificationDetailRequest{
		NotificationId: commonState.Notification.NotificationId,
		TargetId:       s.parentID,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed RetrieveNotificationDetail for notiID %s, targetID %s: %v", commonState.Notification.NotificationId, s.parentID, err)
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

func (s *QuestionaireViewSuite) parentAnswerQuestionnaireForThemself(ctx context.Context) (context.Context, error) {
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

func (s *QuestionaireViewSuite) answerQuestionnaireForThemself(ctx context.Context, person string) (context.Context, error) {
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

func (s *QuestionaireViewSuite) getMapUserIDAndUserGroup(state *common.StepState, person string) (map[string]string, error) {
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

package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
)

// nolint
type QuestionaireCreateSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitQuestionnaireCreate(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &QuestionaireCreateSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^admin upsert notification with updated questionnaire$`:                                                                    s.adminUpsertNotificationWithUpdatedQuestionnaire,
		`^questions with order_index "([^"]*)" are soft deleted$`:                                                                   s.questionsWithOrderIndexAreSoftDeleted,
		`^update questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                                  s.updateQuestionnaireWithResubmitAllowedQuestionsRespectively,
		`^fill all questionnaire_id and questionnaire_question_id from db into payload`:                                             s.fillAllQuestionnaireIDAndQuestionnaireQuestionIDFromDBInfoPayload,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff create a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                  s.CurrentStaffCreateQuestionnaire,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^questionnaire and qn_question are correctly stored in db$`:                                                                   s.questionnaireAndQNQuestionAreCorrectlyStoredInDB,
		`^school admin add packages data of those courses for each student$`:                                                           s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^school admin creates "([^"]*)" course$`:                                                                                      s.CreatesNumberOfCourses,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents$`:                                                             s.CreatesNumberOfStudentsWithParentsInfo,
		`^admin remove questionnaire and upsert notification again$`:                                                                   s.adminRemoveQuestionnaireAndUpsertNotificationAgain,
		`^notification has no questionnaire in DB$`:                                                                                    s.notificationHasNoQuestionnaireInDB,
		`^questionnaire and qn_question are soft deleted in db$`:                                                                       s.questionnaireAndQNQuestionAreSoftDeletedInDB,
		`^current staff discards notification$`:                                                                                        s.CurrentStaffDiscardsNotification,
		`^notification is discarded$`:                                                                                                  s.NotificationIsDiscarded,
		`^current staff create questionnaire template from questionnaire$`:                                                             s.currentStaffCreateQuestionnaireTemplate,
		`^current staff update with updated questionnaire template with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`: s.updateWithUpdatedQuestionnaireTemplateWithResubmittedAndQuestion,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *QuestionaireCreateSuite) updateQuestionnaireWithResubmitAllowedQuestionsRespectively(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	questions := s.CommunicationHelper.ParseQuestionFromString(questionStr)

	for _, questionNew := range questions {
		for _, questionOld := range commonState.Questionnaire.Questions {
			if questionNew.OrderIndex == questionOld.OrderIndex {
				questionNew.QuestionnaireQuestionId = questionOld.QuestionnaireQuestionId
			}
		}
	}

	commonState.Questionnaire.ResubmitAllowed = common.StrToBool(resubmit)
	commonState.Questionnaire.Questions = questions
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionaireCreateSuite) questionsWithOrderIndexAreSoftDeleted(ctx context.Context, orderIdxesStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	orderIdxes := make([]string, 0)
	if orderIdxesStr != "" {
		orderIdxes = strings.Split(orderIdxesStr, ",")
	}
	idxes := make([]int32, 0, len(orderIdxes))
	for _, item := range orderIdxes {
		idx, err := strconv.Atoi(item)
		if err != nil {
			return ctx, fmt.Errorf("convert err: %v", err)
		}
		idxes = append(idxes, int32(idx))
	}
	query := `
	select count(*) from questionnaire_questions qq join info_notifications ins using(questionnaire_id) where qq.deleted_at is not null
	and ins.notification_id=$1 and qq.order_index=ANY($2)`
	var count pgtype.Int8
	err := s.BobDBConn.QueryRow(ctx, query, commonState.Notification.NotificationId, database.Int4Array(idxes)).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("s.Connections.BobDB.QueryRow: %v", err)
	}
	if int(count.Int) != len(orderIdxes) {
		return ctx, fmt.Errorf("quesions %v are not fully soft deleted", orderIdxes)
	}
	return ctx, nil
}

func (s *QuestionaireCreateSuite) adminUpsertNotificationWithUpdatedQuestionnaire(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	upsert := &npb.UpsertNotificationRequest{
		Notification:  commonState.Notification,
		Questionnaire: commonState.Questionnaire,
	}

	_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(ctx, upsert)
	return ctx, err
}

func (s *QuestionaireCreateSuite) questionnaireAndQNQuestionAreSoftDeletedInDB(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	qnID := commonState.Questionnaire.QuestionnaireId
	var count pgtype.Int8
	err := s.BobDBConn.QueryRow(ctx, `select count(*) from questionnaires where questionnaire_id = $1 and deleted_at is not null`, qnID).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("error when get count questionnaire in db, questionnaire_id %s: %v", qnID, err)
	}
	if count.Int != 1 {
		return ctx, fmt.Errorf("counting softdeleted questionnaire with id %s return %d", qnID, count.Int)
	}
	return ctx, nil
}

func (s *QuestionaireCreateSuite) notificationHasNoQuestionnaireInDB(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notiID := commonState.Notification.NotificationId
	var count pgtype.Int8
	err := s.BobDBConn.QueryRow(ctx, `select count(*) from info_notifications where notification_id = $1 and questionnaire_id is null`, notiID).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("error when get count info_notifications in db, notification_id %s: %v", notiID, err)
	}
	if count.Int != 1 {
		return ctx, fmt.Errorf("counting noti with id %s and null questionnaire return %d", notiID, count.Int)
	}
	return ctx, nil
}

func (s *QuestionaireCreateSuite) adminRemoveQuestionnaireAndUpsertNotificationAgain(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	req := &npb.UpsertNotificationRequest{
		Notification: commonState.Notification,
	}

	_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(ctx, req)
	return ctx, err
}

func (s *QuestionaireCreateSuite) fillAllQuestionnaireIDAndQuestionnaireQuestionIDFromDBInfoPayload(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notiID := commonState.Notification.NotificationId
	questionnaire := commonState.Questionnaire

	for index, question := range questionnaire.Questions {
		query := `
		select questionnaire_question_id from questionnaire_questions qq join info_notifications ins
		using (questionnaire_id) where ins.notification_id=$1 and qq.order_index = $2 limit 1`

		row := s.BobDBConn.QueryRow(ctx, query, notiID, question.OrderIndex)

		var questionnaireQuestionID string
		err := row.Scan(&questionnaireQuestionID)
		if err != nil {
			return ctx, err
		}

		questionnaire.Questions[index].QuestionnaireQuestionId = questionnaireQuestionID
	}

	return common.StepStateToContext(ctx, commonState), nil
}
func (s *QuestionaireCreateSuite) questionnaireAndQNQuestionAreCorrectlyStoredInDB(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notiID := commonState.Notification.NotificationId
	query := `
	select questionnaire_id, order_index, qq.type, title, is_required from questionnaire_questions qq join info_notifications ins
	using (questionnaire_id) where ins.notification_id=$1 and qq.deleted_at is null
	`
	rows, err := s.BobDBConn.Query(ctx, query, notiID)
	if err != nil {
		return ctx, err
	}
	defer rows.Close()
	var count int
	expectQuestions := commonState.Questionnaire.Questions
	checkList := makeQuestionCheckList(expectQuestions)
	for rows.Next() {
		count++
		var ent entities.QuestionnaireQuestion
		err = rows.Scan(&ent.QuestionnaireID, &ent.OrderIndex, &ent.Type, &ent.Title, &ent.IsRequired)
		if err != nil {
			return ctx, err
		}
		matched, exist := checkList[int64(ent.OrderIndex.Int)]
		if !exist {
			return ctx, fmt.Errorf("missing question with order idx %d, notiID %s", ent.OrderIndex.Int, notiID)
		}
		delete(checkList, int64(ent.OrderIndex.Int))
		if matched.Title != ent.Title.String {
			return ctx, fmt.Errorf("expected title: %s, got %s", matched.Title, ent.Title.String)
		}
		if matched.Required != ent.IsRequired.Bool {
			return ctx, fmt.Errorf("expected required is: %v, got %v", matched.Required, ent.IsRequired.Bool)
		}
		if matched.Type.String() != ent.Type.String {
			return ctx, fmt.Errorf("expected type is: %s, got %s", matched.Type.String(), ent.Type.String)
		}
	}
	if len(checkList) != 0 {
		return ctx, fmt.Errorf("these questions are missed querying db %v", checkList)
	}
	return ctx, nil
}

func makeQuestionCheckList(questions []*cpb.Question) map[int64]*cpb.Question {
	ret := map[int64]*cpb.Question{}
	for _, item := range questions {
		ret[item.OrderIndex] = item
	}
	return ret
}

func (s *QuestionaireCreateSuite) currentStaffCreateQuestionnaireTemplate(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	questionnaire := commonState.Questionnaire

	svc := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn)

	questionnaireTemplate := &npb.QuestionnaireTemplate{
		QuestionnaireTemplateId: idutil.ULIDNow(),
		Name:                    idutil.ULIDNow(),
		ResubmitAllowed:         questionnaire.ResubmitAllowed,
		ExpirationDate:          questionnaire.ExpirationDate,
		Questions:               s.CommunicationHelper.MapQuestionnaireQuestionToQuestionnaireTemplateQuestion(questionnaire.Questions),
	}

	commonState.QuestionnaireTemplate = questionnaireTemplate

	res, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
		QuestionnaireTemplate: questionnaireTemplate,
	})
	if err != nil {
		return ctx, fmt.Errorf("questionnaire template created failed: %v", err)
	}

	commonState.Questionnaire.QuestionnaireTemplateId = res.QuestionnaireTemplateId

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionaireCreateSuite) updateWithUpdatedQuestionnaireTemplateWithResubmittedAndQuestion(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	questions := s.CommunicationHelper.ParseQuestionTemplateFromString(questionStr)

	for _, questionNew := range questions {
		for _, questionOld := range commonState.QuestionnaireTemplate.Questions {
			if questionNew.OrderIndex == questionOld.OrderIndex {
				questionNew.QuestionnaireTemplateQuestionId = questionOld.QuestionnaireTemplateQuestionId
			}
		}
	}

	commonState.QuestionnaireTemplate.ResubmitAllowed = common.StrToBool(resubmit)
	commonState.QuestionnaireTemplate.Questions = questions

	svc := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn)

	res, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
		QuestionnaireTemplate: commonState.QuestionnaireTemplate,
	})
	if err != nil {
		return ctx, fmt.Errorf("questionnaire template created failed: %v", err)
	}

	commonState.Questionnaire.QuestionnaireTemplateId = res.QuestionnaireTemplateId

	return common.StepStateToContext(ctx, commonState), nil
}

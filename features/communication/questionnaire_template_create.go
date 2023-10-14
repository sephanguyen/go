package communication

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuestionnaireTemplateCreateSuite struct {
	*common.NotificationSuite
	existError error
}

func (c *SuiteConstructor) InitQuestionnaireTemplateCreate(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &QuestionnaireTemplateCreateSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff create a questionnaire template with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:         s.currentStaffCreateQuestionnaireTemplate,
		`^questionnaire template and question are correctly stored in db$`:                                                          s.questionnaireTemplateAndQuestionAreCorrectlyStoredInDB,
		`^current staff update questionnaire template with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:           s.currentStaffUpdateQuestionnaireTemplate,
		`^questions with order_index "([^"]*)" are soft deleted$`:                                                                   s.questionsWithOrderIndexAreSoftDeleted,
		`^current staff create questionnaire template with name is exist$`:                                                          s.createQuestionnaireTemplateWithNameIsExist,
		`^return error questionnaire template name existed$`:                                                                        s.returnErrorQuestionnaireTemplateNameIsExisted,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *QuestionnaireTemplateCreateSuite) currentStaffCreateQuestionnaireTemplate(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	svc := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn)

	questionnaireTemplate := &npb.QuestionnaireTemplate{
		QuestionnaireTemplateId: idutil.ULIDNow(),
		Name:                    idutil.ULIDNow(),
		ResubmitAllowed:         common.StrToBool(resubmit),
		ExpirationDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
		Questions:               s.CommunicationHelper.ParseQuestionTemplateFromString(questionStr),
	}

	commonState.QuestionnaireTemplate = questionnaireTemplate

	_, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
		QuestionnaireTemplate: questionnaireTemplate,
	})

	if err != nil {
		return ctx, fmt.Errorf("questionnaire template created failed: %v", err)
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireTemplateCreateSuite) questionnaireTemplateAndQuestionAreCorrectlyStoredInDB(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	questionnaireTemplate := commonState.QuestionnaireTemplate

	err = s.checkQuestionnaireTemplateInDB(ctx, questionnaireTemplate)
	if err != nil {
		return ctx, err
	}

	err = s.checkQuestionnaireTemplateQuestionInDB(ctx, questionnaireTemplate)
	if err != nil {
		return ctx, err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireTemplateCreateSuite) checkQuestionnaireTemplateInDB(ctx context.Context, questionnaireTemplate *npb.QuestionnaireTemplate) error {
	query := `
		SELECT qt.questionnaire_template_id, qt.name, qt.resubmit_allowed, qt.expiration_date
		FROM public.questionnaire_templates qt
		WHERE qt.questionnaire_template_id = $1 AND qt.deleted_at IS NULL;
	`

	rows, err := s.BobDBConn.Query(ctx, query, questionnaireTemplate.QuestionnaireTemplateId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ent entities.QuestionnaireTemplate
		err = rows.Scan(&ent.QuestionnaireTemplateID, &ent.Name, &ent.ResubmitAllowed, &ent.ExpirationDate)
		if err != nil {
			return err
		}

		if ent.QuestionnaireTemplateID.String != questionnaireTemplate.QuestionnaireTemplateId {
			return fmt.Errorf("expected QuestionnaireTemplateID %s, got %s", questionnaireTemplate.QuestionnaireTemplateId, ent.QuestionnaireTemplateID.String)
		}

		if ent.Name.String != questionnaireTemplate.Name {
			return fmt.Errorf("expected Name %s, got %s", questionnaireTemplate.Name, ent.Name.String)
		}

		if ent.ResubmitAllowed.Bool != questionnaireTemplate.ResubmitAllowed {
			return fmt.Errorf("expected ResubmitAllowed %t, got %t", questionnaireTemplate.ResubmitAllowed, ent.ResubmitAllowed.Bool)
		}
	}

	return nil
}

func (s *QuestionnaireTemplateCreateSuite) checkQuestionnaireTemplateQuestionInDB(ctx context.Context, questionnaireTemplate *npb.QuestionnaireTemplate) error {
	query := `
		SELECT qtq.questionnaire_template_question_id, qtq.title, qtq.order_index, qtq.type, qtq.is_required 
		FROM questionnaire_template_questions qtq
		WHERE qtq.questionnaire_template_id = $1 AND qtq.deleted_at IS NULL;
	`

	rows, err := s.BobDBConn.Query(ctx, query, questionnaireTemplate.QuestionnaireTemplateId)
	if err != nil {
		return err
	}
	defer rows.Close()

	expectQuestions := questionnaireTemplate.Questions
	checkList := makeQuestionnaireTemplateQuestionCheckList(expectQuestions)

	for rows.Next() {
		var ent entities.QuestionnaireTemplateQuestion
		err = rows.Scan(&ent.QuestionnaireTemplateQuestionID, &ent.Title, &ent.OrderIndex, &ent.Type, &ent.IsRequired)
		if err != nil {
			return err
		}

		matched, exist := checkList[int64(ent.OrderIndex.Int)]
		if !exist {
			return fmt.Errorf("missing question with order idx %d, questionnaire template id %s", ent.OrderIndex.Int, questionnaireTemplate.QuestionnaireTemplateId)
		}

		if ent.Title.String != matched.Title {
			return fmt.Errorf("expected Title %s, got %s", matched.Title, ent.Title.String)
		}

		if ent.OrderIndex.Int != int32(matched.OrderIndex) {
			return fmt.Errorf("expected OrderIndex %d, got %d", matched.OrderIndex, ent.OrderIndex.Int)
		}

		if ent.Type.String != matched.Type.String() {
			return fmt.Errorf("expected Type %s, got %s", matched.Type.String(), ent.Type.String)
		}

		if ent.IsRequired.Bool != matched.Required {
			return fmt.Errorf("expected IsRequired %t, got %t", matched.Required, ent.IsRequired.Bool)
		}
	}

	return nil
}

func makeQuestionnaireTemplateQuestionCheckList(questions []*npb.QuestionnaireTemplateQuestion) map[int64]*npb.QuestionnaireTemplateQuestion {
	ret := map[int64]*npb.QuestionnaireTemplateQuestion{}
	for _, item := range questions {
		ret[item.OrderIndex] = item
	}
	return ret
}

func (s *QuestionnaireTemplateCreateSuite) currentStaffUpdateQuestionnaireTemplate(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
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

	_, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
		QuestionnaireTemplate: commonState.QuestionnaireTemplate,
	})

	if err != nil {
		return ctx, fmt.Errorf("questionnaire template created failed: %v", err)
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireTemplateCreateSuite) questionsWithOrderIndexAreSoftDeleted(ctx context.Context, orderIdxesStr string) (context.Context, error) {
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
		idxes = append(idxes, int32(idx)) // #nosec G109
	}

	query := `
		SELECT count(*) FROM questionnaire_template_questions qtq 
		WHERE questionnaire_template_id = $1
		AND qtq.deleted_at IS NOT NULL AND qtq.order_index = ANY($2)
	`
	var count pgtype.Int8
	err := s.BobDBConn.QueryRow(ctx, query, commonState.QuestionnaireTemplate.QuestionnaireTemplateId, database.Int4Array(idxes)).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("s.Connections.BobDB.QueryRow: %v", err)
	}
	if int(count.Int) != len(orderIdxes) {
		return ctx, fmt.Errorf("questions %v are not fully soft deleted", orderIdxes)
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireTemplateCreateSuite) createQuestionnaireTemplateWithNameIsExist(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	svc := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn)

	_, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
		QuestionnaireTemplate: &npb.QuestionnaireTemplate{
			QuestionnaireTemplateId: idutil.ULIDNow(),
			Name:                    commonState.QuestionnaireTemplate.Name,
			ResubmitAllowed:         commonState.QuestionnaireTemplate.ResubmitAllowed,
			ExpirationDate:          commonState.QuestionnaireTemplate.ExpirationDate,
			Questions:               commonState.QuestionnaireTemplate.Questions,
		},
	})

	if err != nil {
		if errors.Is(err, status.Error(codes.InvalidArgument, "questionnaire template name is exist")) {
			s.existError = err
			return ctx, nil
		}
		return ctx, fmt.Errorf("questionnaire template created failed: %v", err)
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *QuestionnaireTemplateCreateSuite) returnErrorQuestionnaireTemplateNameIsExisted(ctx context.Context) (context.Context, error) {
	if s.existError == nil {
		return ctx, fmt.Errorf("returnErrorQuestionnaireTemplateNameIsExisted: expected %v, got %v", status.Error(codes.InvalidArgument, "questionnaire template name is exist"), nil)
	}
	return ctx, nil
}

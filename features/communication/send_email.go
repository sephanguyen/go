package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/cucumber/godog"
)

type SendEmailSuite struct {
	//Todo: rename it to CommunicationSuite
	*common.NotificationSuite
	email   *spb.SendEmailRequest
	emailID string
}

func (c *SuiteConstructor) InitSendEmail(dep *DependencyV2, ctx *godog.ScenarioContext) {
	s := &SendEmailSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^waiting for kafka sync user info for notificationmgmt database$`:                                                          s.waitingForKafkaSyncUserInfo,
		`^current staff send an email$`:                             s.staffSendAnEmail,
		`^returns "([^"]*)" status code$`:                           s.CheckReturnStatusCode,
		`^spike service must save this email and email recipients$`: s.spikeServiceMustSaveThisEmailAndEmailRecipients,
	}
	c.InitScenarioStepMapping(ctx, stepsMapping)
}

func (s *SendEmailSuite) staffSendAnEmail(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	sendEmailReq := &spb.SendEmailRequest{
		Subject: "Just test",
		Content: &spb.SendEmailRequest_EmailContent{
			HTML:      "<h1>Just test.</h1>",
			PlainText: "Just test.",
		},
		Recipients:     []string{},
		OrganizationId: commonState.CurrentResourcePath,
	}

	randNum := common.RandRangeIn(4, 10)
	for i := 0; i < randNum; i++ {
		sendEmailReq.Recipients = append(sendEmailReq.Recipients, fmt.Sprintf("example_recipient_%s@manabie.com", idutil.ULIDNow()))
	}

	emptyCtx := context.Background()

	s.email = sendEmailReq
	commonState.Request = sendEmailReq
	resp, err := spb.NewEmailModifierServiceClient(s.SpikeGRPCConn).SendEmail(emptyCtx, commonState.Request.(*spb.SendEmailRequest))
	commonState.Response, commonState.ResponseErr = resp, err
	s.emailID = resp.EmailId

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendEmailSuite) spikeServiceMustSaveThisEmailAndEmailRecipients(ctx context.Context) (context.Context, error) {
	query := `
		SELECT count(*)
		FROM email_recipients er
		WHERE er.email_id = $1;
	`
	var count int
	err := s.NotificationMgmtDBConn.QueryRow(ctx, query, s.emailID).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("failed query: %v", err)
	}

	if count != len(s.email.Recipients) {
		return ctx, fmt.Errorf("expected email recipients %v, got %v", len(s.email.Recipients), count)
	}
	return ctx, nil
}

func (s *SendEmailSuite) waitingForKafkaSyncUserInfo(ctx context.Context) (context.Context, error) {
	fmt.Printf("Waiting for kafka sync user info from bob to notificationmgmt database...")
	time.Sleep(3 * time.Second)
	return ctx, nil
}

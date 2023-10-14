package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type SetStatusForSystemNotificationSuite struct {
	*common.NotificationSuite
	markedSystemNotificationID string
}

func (c *SuiteConstructor) InitSetStatusForSystemNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SetStatusForSystemNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^some staffs with random roles and granted organization location of current organization$`:                                 s.CreateSomeStaffsWithSomeRolesAndGrantedOrgLevelLocationOfCurrentOrganization,
		`^user set "([^"]*)" the system notification$`:                                                                              s.userSetStatusTheSystemNotification,
		`^mark the system notification as status "([^"]*)"$`:                                                                        s.markTheSystemNotificationAsStatus,
		`^waiting for kafka sync data$`:                                                                                             s.waitingForKafkaSync,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^staff create system notification with "([^"]*)" new and "([^"]*)" done and "([^"]*)" unenabled$`:                          s.CreateNumberOfSystemNotificationWithSomeStatus,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SetStatusForSystemNotificationSuite) waitingForKafkaSync(ctx context.Context) (context.Context, error) {
	fmt.Printf("Waiting for kafka sync data...\n")
	time.Sleep(10 * time.Second)
	return ctx, nil
}

func (s *SetStatusForSystemNotificationSuite) userSetStatusTheSystemNotification(ctx context.Context, status string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	systemNotificationsList, err := s.GetSystemNotificationByUser(commonState.TokenOfSentRecipient)
	if err != nil {
		return ctx, fmt.Errorf("failed GetSystemNotificationByUser: %+v", err)
	}

	s.markedSystemNotificationID = systemNotificationsList[0].GetSystemNotificationId()
	req := &npb.SetSystemNotificationStatusRequest{
		SystemNotificationId: s.markedSystemNotificationID,
	}
	switch status {
	case npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String():
		req.Status = npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE
	case npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String():
		req.Status = npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW
	}

	ctx2, cancel := common.ContextWithTokenAndTimeOut(ctx, commonState.TokenOfSentRecipient)
	defer cancel()
	_, err = npb.NewSystemNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SetSystemNotificationStatus(ctx2, req)

	if err != nil {
		return ctx, fmt.Errorf("SetSystemNotificationStatus: %v", err)
	}

	return ctx, nil
}

func (s *SetStatusForSystemNotificationSuite) markTheSystemNotificationAsStatus(ctx context.Context, expectStatus string) (context.Context, error) {
	query := `
		SELECT status
		FROM system_notifications sn
		WHERE sn.system_notification_id = $1
	`
	var status string
	err := s.NotificationMgmtPostgresDBConn.QueryRow(ctx, query, database.Text(s.markedSystemNotificationID)).Scan(&status)
	if err != nil {
		return ctx, fmt.Errorf("failed QueryRow: %+v", err)
	}

	if status != expectStatus {
		return ctx, fmt.Errorf("expected status %s , got %s", expectStatus, status)
	}
	return ctx, nil
}

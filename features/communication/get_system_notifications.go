package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type GetSystemNotificationsSuite struct {
	*common.NotificationSuite
	systemNotifications *npb.RetrieveSystemNotificationsResponse
}

func (c *SuiteConstructor) InitGetSystemNotifications(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &GetSystemNotificationsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^staff get system notifications with status "([^"]*)" and lang "([^"]*)" and limit "([^"]*)" and offset "([^"]*)"$`:        s.currentStaffGetSystemNotificationsByPagingWithLimitAndOffset,
		`^waiting for kafka sync data$`: s.waitingForKafkaSync,
		`^some staffs with random roles and granted organization location of current organization$`:        s.CreateSomeStaffsWithSomeRolesAndGrantedOrgLevelLocationOfCurrentOrganization,
		`^staff create system notification with "([^"]*)" new and "([^"]*)" done and "([^"]*)" unenabled$`: s.CreateNumberOfSystemNotificationWithSomeStatus,
		`^staff check response "([^"]*)" new and "([^"]*)" done and "([^"]*)" status and "([^"]*)" count$`: s.staffCheckResponseForSystemNotificationsWithStatus,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetSystemNotificationsSuite) waitingForKafkaSync(ctx context.Context) (context.Context, error) {
	fmt.Printf("Waiting for kafka sync data...\n")
	time.Sleep(10 * time.Second)
	return ctx, nil
}

func (s *GetSystemNotificationsSuite) currentStaffGetSystemNotificationsByPagingWithLimitAndOffset(ctx context.Context, status, lang, limit, offset string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	var limitReq = uint32(0)
	var offsetReq = uint64(0)
	if limit != "none" {
		limitReq = uint32(strToI32(limit))
	}
	if offset != "none" {
		offsetReq = uint64(strToI32(offset))
	}

	req := &npb.RetrieveSystemNotificationsRequest{
		Paging: &cpb.Paging{
			Limit: limitReq,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(offsetReq),
			},
		},
		Language: lang,
	}

	if status == "new" {
		req.Status = npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW
	} else if status == "done" {
		req.Status = npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE
	}

	res, err := npb.NewSystemNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).
		RetrieveSystemNotifications(s.ContextWithToken(ctx, stepState.TokenOfSentRecipient), req)

	if err != nil {
		return ctx, fmt.Errorf("failed to get system notifications: %v", err)
	}
	s.systemNotifications = res

	return ctx, nil
}

func (s *GetSystemNotificationsSuite) staffCheckResponseForSystemNotificationsWithStatus(ctx context.Context, numNew, numDone, status, totalCount string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	systemNotification := stepState.PayloadSystemNotifications
	mapExpected := make(map[string]*payload.UpsertSystemNotification, len(systemNotification))
	for _, e := range systemNotification {
		mapExpected[e.ReferenceID] = e
	}

	actualSystemNotification := s.systemNotifications.SystemNotifications

	for i := 0; i < len(actualSystemNotification); i++ {
		// check for item returned with correct ValidFrom <= now()
		if actualSystemNotification[i].ValidFrom.AsTime().After(time.Now()) {
			return ctx, fmt.Errorf("error System Notification stored incorrect ValidFrom data. reference ID: %+v", actualSystemNotification[i].Content)
		}

		if expect, exist := mapExpected[actualSystemNotification[i].Url]; exist {
			mapExpectContentValue := make(map[string]string, 0)
			for _, content := range expect.Content {
				mapExpectContentValue[content.Language] = content.Text
			}
			for _, actualContent := range actualSystemNotification[i].GetContent() {
				if value, found := mapExpectContentValue[actualContent.Language]; found {
					if value != actualContent.Text {
						return ctx, fmt.Errorf("incorrect content value, expected %s, found %s. reference ID: %+v", value, actualContent.Text, expect.ReferenceID)
					}
				}
			}

			if expect.URL != actualSystemNotification[i].Url {
				return ctx, fmt.Errorf("error System Notification stored incorrect URL data. reference ID: %+v", expect.ReferenceID)
			}

			if (status == "all" && expect.Status != payload.SystemNotificationStatus(actualSystemNotification[i].Status.String())) ||
				(status == "new" && expect.Status != payload.SystemNotificationStatusNew) ||
				(status == "done" && expect.Status != payload.SystemNotificationStatusDone) {
				return ctx, fmt.Errorf("error System Notification stored incorrect Status data. reference ID: %+v", expect.ReferenceID)
			}

			// check for item returned with correct ValidFrom as expected
			if !expect.ValidFrom.Truncate(time.Millisecond).Equal(actualSystemNotification[i].ValidFrom.AsTime().Truncate(time.Millisecond)) {
				return ctx, fmt.Errorf("error System Notification stored incorrect ValidFrom data. reference ID: %+v", expect.ReferenceID)
			}
		}
	}

	totalItemsForStatus := s.systemNotifications.TotalItemsForStatus
	numSystemNotificationNew := common.StrToInt(numNew)
	numSystemNotificationDone := common.StrToInt(numDone)
	numSystemNotificationAll := common.StrToInt(totalCount)

	if s.systemNotifications.TotalItems != uint32(numSystemNotificationAll) {
		return ctx, fmt.Errorf("error System Notification totalItems")
	}

	for i := 0; i < len(totalItemsForStatus); i++ {
		if totalItemsForStatus[i].Status == npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE && totalItemsForStatus[i].TotalItems != uint32(numSystemNotificationAll) {
			return ctx, fmt.Errorf("error System Notification count for SYSTEM_NOTIFICATION_STATUS_NONE")
		}
		if totalItemsForStatus[i].Status == npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW && totalItemsForStatus[i].TotalItems != uint32(numSystemNotificationNew) {
			return ctx, fmt.Errorf("error System Notification count for SYSTEM_NOTIFICATION_STATUS_NEW")
		}
		if totalItemsForStatus[i].Status == npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE && totalItemsForStatus[i].TotalItems != uint32(numSystemNotificationDone) {
			return ctx, fmt.Errorf("error System Notification count for SYSTEM_NOTIFICATION_STATUS_DONE")
		}
	}

	return ctx, nil
}

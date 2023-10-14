package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	notificationRepo "github.com/manabie-com/backend/internal/notification/repositories"

	"github.com/cucumber/godog"
)

type SendScheduledNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitSendScheduledNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendScheduledNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                             s.CheckReturnStatusCode,
		`^current staff send notification$`:                                           s.CurrentStaffSendNotification,
		`^current staff send notification again$`:                                     s.CurrentStaffSendNotification,
		`^notificationmgmt services must send notification to user$`:                  s.NotificationMgmtMustSendNotificationToUser,
		`^update user device token to an "([^"]*)" device token$`:                     s.UpdateDeviceTokenForLeanerUser,
		`^current staff discards notification$`:                                       s.CurrentStaffDiscardsNotification,
		`^notification is discarded$`:                                                 s.NotificationIsDiscarded,
		`^returns error message "([^"]*)"$`:                                           s.CheckReturnsErrorMessage,
		`^notificationmgmt services must store the notification with correctly info$`: s.NotificationMgmtMustStoreTheNotification,
		`^recipient must receive the notification through FCM mock$`:                  s.recipientMustReceiveTheNotification,
		`^wait for FCM is sent to target user$`:                                       s.WaitingForFCMIsSent,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendScheduledNotificationSuite) recipientMustReceiveTheNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	userNotifications, err := s.GetSendUsersFromDB(ctx)
	if err != nil {
		return ctx, err
	}

	userDeviceTokenRepo := &notificationRepo.UserDeviceTokenRepo{}
	for _, un := range userNotifications {
		var deviceToken string
		receiverID := un.UserID.String
		userDeviceTokens, err := userDeviceTokenRepo.FindByUserIDs(ctx, s.BobDBConn, database.TextArray([]string{receiverID}))
		if err != nil {
			return ctx, fmt.Errorf("error FindByUserIDs %w", err)
		}
		if len(userDeviceTokens) != 1 {
			return ctx, fmt.Errorf("FindByUserIDs expected %d result, got %d", 1, len(userDeviceTokens))
		}
		if userDeviceTokens[0].UserID.String != receiverID {
			return ctx, fmt.Errorf("FindByUserIDs epected user id %s, got %s", receiverID, userDeviceTokens[0].UserID.String)
		}

		deviceToken = userDeviceTokens[0].DeviceToken.String
		resp, err := retrievePushedNotification(ctx, s.NotificationMgmtGRPCConn, deviceToken)
		if err != nil {
			return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
		}
		if len(resp.Messages) == 0 {
			return ctx, fmt.Errorf("err: user receiver id %s, device token %s don't receive any notification", receiverID, deviceToken)
		}

		gotNoti := resp.Messages[len(resp.Messages)-1]
		gotTile := gotNoti.Title
		if gotTile != commonState.Notification.Message.Title {
			return ctx, fmt.Errorf("want notification title to be: %s, got %s", commonState.Notification.Message.Title, gotTile)
		}
	}

	return ctx, nil
}

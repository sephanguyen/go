package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
)

type SendNotificationWithFcmStepIsFailedSuite struct {
	*common.NotificationSuite
	NotificationNeedToSent *cpb.Notification
}

func (c *SuiteConstructor) InitSendNotificationWithFcmStepIsFailed(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendNotificationWithFcmStepIsFailedSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                                  s.CheckReturnStatusCode,
		`^notificationmgmt services must store the notification with correctly info$`:      s.NotificationMgmtMustStoreTheNotification,
		`^notificationmgmt services must send notification to user$`:                       s.NotificationMgmtMustSendNotificationToUser,
		`^recipient must not receive the notification through FCM mock$`:                   s.recipientMustNotReceiveTheNotificationThroughFcmMock,
		`^update user device token to an "([^"]*)" device token$`:                          s.UpdateDeviceTokenForLeanerUser,
		`^update user device token to an "([^"]*)" device token with "([^"]*)" fail rate$`: s.UpdateDeviceTokenForLeanerUserWithPercentageFailure,
		`^current staff send notification$`:                                                s.CurrentStaffSendNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendNotificationWithFcmStepIsFailedSuite) recipientMustNotReceiveTheNotificationThroughFcmMock(ctx context.Context) (context.Context, error) {
	userNotifications, err := s.GetSendUsersFromDB(ctx)
	if err != nil {
		return ctx, err
	}

	for _, un := range userNotifications {
		var deviceToken string
		receiveID := un.UserID.String
		row := s.BobDBConn.QueryRow(ctx, "SELECT device_token FROM public.user_device_tokens WHERE user_id =  $1", receiveID)

		if err := row.Scan(&deviceToken); err != nil {
			return ctx, fmt.Errorf("error finding user device token with userid: %s: %w", receiveID, err)
		}

		resp, err := retrievePushedNotification(ctx, s.NotificationMgmtGRPCConn, deviceToken)
		if err != nil {
			return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
		}
		if len(resp.Messages) != 0 {
			// if receiver's device token is Invalid then they must not receive pushed notification
			if strings.Contains(deviceToken, "invalid") {
				return ctx, fmt.Errorf("expected 0 notification form FCM mock, but got %s, userID %s", strconv.Itoa(len(resp.Messages)), receiveID)
			}
		}
	}

	return ctx, nil
}

package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type SendNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitSendNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                 s.CheckReturnStatusCode,
		`^current staff send notification$`:                               s.CurrentStaffSendNotification,
		`^notificationmgmt services must send notification to user$`:      s.NotificationMgmtMustSendNotificationToUser,
		`^update user device token to an "([^"]*)" device token$`:         s.UpdateDeviceTokenForLeanerUser,
		`^recipient must receive the notification through FCM mock$`:      s.recipientMustReceiveTheNotification,
		`^current staff upsert "([^"]*)" notification missing "([^"]*)"$`: s.upsertNotificationMissing,
		`^username is saved follow by their notification$`:                s.usernameIsSavedFollowByTheirNotification,
		`^wait for FCM is sent to target user$`:                           s.WaitingForFCMIsSent,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendNotificationSuite) recipientMustReceiveTheNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	userNotifications, err := s.GetSendUsersFromDB(ctx)
	if err != nil {
		return ctx, err
	}

	userDeviceTokenRepo := &repositories.UserDeviceTokenRepo{}
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

func (s *SendNotificationSuite) upsertNotificationMissing(ctx context.Context, status, missedField string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	opts := &common.NotificationWithOpts{
		UserGroups:       "student",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "random",
		ScheduledStatus:  "random",
		Status:           "NOTIFICATION_STATUS_SCHEDULED",
		IsImportant:      false,
	}
	var err error
	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}

	switch missedField {
	case "title":
		commonState.Notification.Message.Title = ""
	case "content":
		commonState.Notification.Message.Content = &cpb.RichText{
			Raw:      "",
			Rendered: "",
		}
	}

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: commonState.Notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))
	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		commonState.Notification.NotificationId = resp.NotificationId
	} else {
		return common.StepStateToContext(ctx, commonState), err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendNotificationSuite) usernameIsSavedFollowByTheirNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	userNotificationRepo := repositories.UsersInfoNotificationRepo{}
	userRepo := repositories.UserRepo{}
	filter := repositories.NewFindUserNotificationFilter()
	filter.NotiIDs = database.TextArray([]string{commonState.Notification.NotificationId})

	userNotifications, err := userNotificationRepo.Find(ctx, s.BobDBConn, filter)
	if err != nil {
		return ctx, fmt.Errorf("cannot find user notification: %v", err)
	}

	for _, userNotification := range userNotifications {
		userFilter := &repositories.FindUserFilter{
			UserIDs: database.TextArray([]string{userNotification.UserID.String}),
		}
		users, _, err := userRepo.FindUser(ctx, s.BobDBConn, userFilter)
		if err != nil {
			return ctx, fmt.Errorf("cannot find user: %v", err)
		}
		if len(users) == 0 {
			return ctx, fmt.Errorf("cannot find user: user not found")
		}

		switch userNotification.UserGroup.String {
		case cpb.UserGroup_USER_GROUP_STUDENT.String():
			if users[0].Name.String != userNotification.StudentName.String {
				return ctx, fmt.Errorf("student name does't match, expected: %v, got: %v", users[0].Name.String, userNotification.StudentName.String)
			}
		case cpb.UserGroup_USER_GROUP_PARENT.String():
			if users[0].Name.String != userNotification.ParentName.String {
				return ctx, fmt.Errorf("parent name does't match, expected: %v, got: %v", users[0].Name.String, userNotification.StudentName.String)
			}

			if userNotification.StudentID.String != "" {
				userFilter.UserIDs = database.TextArray([]string{userNotification.StudentID.String})
				users, _, err := userRepo.FindUser(ctx, s.BobDBConn, userFilter)
				if err != nil {
					return ctx, fmt.Errorf("cannot find user: %v", err)
				}
				if len(users) == 0 {
					return ctx, fmt.Errorf("cannot find user: user not found")
				}

				if users[0].Name.String != userNotification.StudentName.String {
					return ctx, fmt.Errorf("student name does't match, expected: %v, got: %v", users[0].Name.String, userNotification.StudentName.String)
				}
			}
		}
	}

	return ctx, nil
}

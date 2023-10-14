package communication

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
)

type NotifyUserUnreadNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitNotifyUserUnreadNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &NotifyUserUnreadNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^notificationmgmt services must store the notification with correctly info$`: s.NotificationMgmtMustStoreTheNotification,
		`^current staff send notification$`:                                           s.CurrentStaffSendNotification,
		`^returns "([^"]*)" status code$`:                                             s.CheckReturnStatusCode,
		`^notificationmgmt services must send notification to user$`:                  s.NotificationMgmtMustSendNotificationToUser,
		`^returns error message "([^"]*)"$`:                                           s.CheckReturnsErrorMessage,
		`^current staff notifies notification to unread users$`:                       s.currentStaffNotifiesNotificationToUnreadUsers,
		`^some users read notification$`:                                              s.someUsersReadNotification,
		`^update user device token to an "([^"]*)" device token$`:                     s.UpdateDeviceTokenForLeanerUser,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *NotifyUserUnreadNotificationSuite) currentStaffNotifiesNotificationToUnreadUsers(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).NotifyUnreadUser(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.NotifyUnreadUserRequest{NotificationId: commonState.Notification.NotificationId})

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *NotifyUserUnreadNotificationSuite) someUsersReadNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	userInfoNotificationRepo := &repositories.UsersInfoNotificationRepo{}
	var numOfUser int64 = 100
	filter := repositories.FindUserNotificationFilter{
		UserIDs:     pgtype.TextArray{Status: pgtype.Null},
		NotiIDs:     database.TextArray([]string{commonState.Notification.NotificationId}),
		UserStatus:  pgtype.TextArray{Status: pgtype.Null},
		Limit:       database.Int8(numOfUser),
		OffsetText:  pgtype.Text{Status: pgtype.Null},
		IsImportant: pgtype.Bool{Status: pgtype.Null},
	}

	unsMap, err := userInfoNotificationRepo.FindUserIDs(ctx, s.BobDBConn, filter)
	if err != nil {
		return ctx, err
	}
	uns := unsMap[commonState.Notification.NotificationId]

	readUser := make([]string, 0)
	for _, un := range uns {
		// user read notification or not
		if rand.Intn(2) > 0 {
			readUser = append(readUser, un.UserID.String)
		}
	}

	ctx, err = s.userReadNotification(ctx, readUser, commonState.Notification.NotificationId)
	if err != nil {
		return ctx, err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *NotifyUserUnreadNotificationSuite) userReadNotification(ctx context.Context, userID []string, notificationID string) (context.Context, error) {
	commonState := StepStateFromContext(ctx)
	query := `UPDATE users_info_notifications SET status = $1 WHERE user_id = ANY($2) AND notification_id = ANY($3) AND deleted_at IS NULL;`

	cmd, err := s.BobDBConn.Exec(
		ctx,
		query,
		database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ.String()),
		database.TextArray(userID),
		database.TextArray([]string{notificationID}),
	)
	if err != nil {
		return ctx, err
	}

	if len(userID) > 0 && cmd.RowsAffected() == 0 {
		return ctx, fmt.Errorf("no rows affected")
	}
	return StepStateToContext(ctx, commonState), nil
}

package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type DeleteNotificationSuite struct {
	*common.NotificationSuite
	NotificationNeedToSent *cpb.Notification
}

func (c *SuiteConstructor) InitDeleteNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &DeleteNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^current staff send notification$`:                          s.CurrentStaffSendNotification,
		`^current staff deletes notification$`:                       s.CurrentStaffDeletesNotification,
		`^notification is deleted$`:                                  s.NotificationIsDiscarded,
		`^returns "([^"]*)" status code$`:                            s.CheckReturnStatusCode,
		`^returns error message "([^"]*)"$`:                          s.CheckReturnsErrorMessage,
		`^notificationmgmt services must send notification to user$`: s.NotificationMgmtMustSendNotificationToUser,
		`^current staff create a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`: s.CurrentStaffCreateQuestionnaire,
		`^recipient must not retrieve notification$`:                                                               s.recipientMustNotRetrieveNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *DeleteNotificationSuite) recipientMustNotRetrieveNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	sentNotification := commonState.Request.(*npb.SendNotificationRequest)

	// get a recipient from the list of recipient of that notification
	query := `
	SELECT user_id
	FROM users_info_notifications uin
	WHERE uin.notification_id = $1
	LIMIT 1
	`
	var recipientID string
	err := s.BobDBConn.QueryRow(ctx, query, database.Text(sentNotification.NotificationId)).Scan(&recipientID)
	if err != nil {
		return ctx, fmt.Errorf("failed QueryRow: %+v", err)
	}

	if recipientID == "" {
		return ctx, fmt.Errorf("unexpected empty recipientID. notification %s", sentNotification.NotificationId)
	}

	// find token of the recipientID
	userRole := ""
	for _, student := range commonState.Students {
		if student.ID == recipientID {
			userRole = "student"
			break
		}
		for _, parent := range student.Parents {
			if parent.ID == recipientID {
				userRole = "parent"
				break
			}
		}
	}

	token, err := s.GenerateExchangeTokenCtx(ctx, recipientID, userRole)
	if err != nil {
		return ctx, fmt.Errorf("failed login learner app: %v", err)
	}

	userNotifications, err := s.GetNotificationByUser(token, false)
	if err != nil {
		return ctx, fmt.Errorf("GetUserNotification %s", err)
	}

	if len(userNotifications) > 0 {
		return ctx, fmt.Errorf("expected user %s retrieve no notification", recipientID)
	}

	return ctx, nil
}

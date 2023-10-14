package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"

	"github.com/cucumber/godog"
)

type SendNotificationWithExcludedGenericReceiverIdsSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitSendNotificationWithExcludedGenericReceiverIds(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendNotificationWithExcludedGenericReceiverIdsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^current staff upsert notification with valid filter$`:                                                                     s.upsertNotificationWithValidFilter,
		`^current staff send notification$`:                                                                                         s.CurrentStaffSendNotification,
		`^notificationmgmt services must send notification to user$`:                                                                s.NotificationMgmtMustSendNotificationToUser,
		`^returns "([^"]*)" status code$`:                                                                                           s.CheckReturnStatusCode,
		`^excluded user must not receive notification$`:                                                                             s.excludedUserMustNotReceiveNotification,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendNotificationWithExcludedGenericReceiverIdsSuite) upsertNotificationWithValidFilter(ctx context.Context) (context.Context, error) {
	opts := &common.NotificationWithOpts{
		UserGroups:                 "student",
		CourseFilter:               "all",
		GradeFilter:                "all",
		LocationFilter:             "none",
		ClassFilter:                "none",
		IndividualFilter:           "none",
		ScheduledStatus:            "none",
		Status:                     "NOTIFICATION_STATUS_DRAFT",
		IsImportant:                false,
		ExcludedGenericReceiverStr: "excluded",
	}

	return s.CurrentStaffUpsertNotificationWithOpts(ctx, opts)
}

func (s *SendNotificationWithExcludedGenericReceiverIdsSuite) excludedUserMustNotReceiveNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	query := `
		SELECT count(*)
		FROM users_info_notifications uin
		WHERE uin.notification_id = $1 AND user_id = ANY($2::TEXT[]);
	`
	var count int
	err := s.BobDBConn.QueryRow(ctx, query, commonState.Notification.NotificationId, commonState.Notification.ExcludedGenericReceiverIds).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("failed query: %v", err)
	}

	if count > 0 {
		return ctx, fmt.Errorf("expected notification %v will not send to excluded generic receiver ids", commonState.Notification.NotificationId)
	}

	return ctx, nil
}

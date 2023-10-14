package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
	"golang.org/x/exp/slices"
)

type ParentWithMultipleStudent struct {
	*common.NotificationSuite
	studentIDs  []string
	parentID    string
	parentToken string
}

func (c *SuiteConstructor) InitNotificationForParentWithMultipleStudent(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &ParentWithMultipleStudent{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^admin create notification sending to parent of students$`:                                                                 s.adminCreateNotificationSendingToParentOfStudents,
		`^parent has (\d+) items in notification list$`:                                                                             s.parentHasItemsInNotificationList,
		`^parent has (\d+) unread notification$`:                                                                                    s.parentHasUnreadNotification,
		`^parent read notification using created notification id$`:                                                                  s.parentReadNotificationUsingCreatedNotificationID,
		`^school admin has created (\d+) students with the same parent$`:                                                            s.CreatesNumberOfStudentsWithSameParentsInfo,
		`^parent login to Learner App$`:                                                                                             s.parentLoginToLearnerApp,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *ParentWithMultipleStudent) parentReadNotificationUsingCreatedNotificationID(ctx context.Context) (context.Context, error) {
	return s.UserSetStatusToNotification(ctx, s.parentToken, "USER_NOTIFICATION_STATUS_READ")
}

func (s *ParentWithMultipleStudent) parentHasItemsInNotificationList(ctx context.Context, count int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notis, err := s.GetNotificationByUser(s.parentToken, false)
	if err != nil {
		return ctx, err
	}
	if len(notis) != count {
		return ctx, fmt.Errorf("want %d new notification, got %d, check user_id %s of ctx %s", count, len(notis), s.parentID, s.CurrentResourcePath)
	}
	notiID := commonState.Notification.NotificationId
	// notis must have the same noti_id
	for _, item := range notis {
		if item.UserNotification.NotificationId != notiID {
			return ctx, fmt.Errorf("parent %s does not have correct notification id: %s vs %s", s.parentID, item.UserNotification.NotificationId, notiID)
		}
		if !slices.Contains(s.studentIDs, item.TargetId) {
			return ctx, fmt.Errorf("not expect to receive noti for student %s", item.TargetId)
		}
	}

	return ctx, nil
}

func (s *ParentWithMultipleStudent) parentHasUnreadNotification(ctx context.Context, count int) (context.Context, error) {
	countNotis, err := s.CountUserNotificationByStatus(s.parentToken, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ)
	if err != nil {
		return ctx, err
	}
	countUnread := countNotis.Total - countNotis.NumByStatus
	if countUnread != int32(count) {
		return ctx, fmt.Errorf("want %d new notification, got %d, check user_id %s", count, countUnread, s.parentID)
	}
	return ctx, nil
}

func (s *ParentWithMultipleStudent) adminCreateNotificationSendingToParentOfStudents(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, student := range commonState.Students {
		s.studentIDs = append(s.studentIDs, student.ID)
	}
	opts := &common.NotificationWithOpts{
		UserGroups:       "parent",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "1 min",
		Status:           "NOTIFICATION_STATUS_SCHEDULED",
		IsImportant:      false,
		// ReceiverIds:      s.studentIDs,
	}
	ctx, err := s.CurrentStaffUpsertNotificationWithOpts(ctx, opts)
	if err != nil {
		return ctx, fmt.Errorf("failed upsert notification %v", err)
	}

	ctx, err = s.CurrentStaffSendNotification(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed send notification %v", err)
	}

	return ctx, nil
}
func (s *ParentWithMultipleStudent) parentLoginToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.parentID = commonState.Students[0].Parents[0].ID
	var err error
	s.parentToken, err = s.GenerateExchangeTokenCtx(ctx, s.parentID, commonState.Students[0].Parents[0].Group)
	if err != nil {
		return ctx, nil
	}
	return ctx, nil
}

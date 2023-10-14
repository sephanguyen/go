package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type AccessControlForUpdateNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitAccessControlForUpdateNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &AccessControlForUpdateNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^a new "([^"]*)" and granted "([^"]*)" descendant locations logged in Back Office of a current organization$`:              s.StaffGrantedRoleAndLocationsLoggedInBackOfficeOfCurrentOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                            s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                  s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                              s.CheckReturnStatusCode,
		`^returns "([^"]*)" status code and error message have "([^"]*)"$`:             s.CheckReturnStatusCodeAndContainMsg,
		`^notificationmgmt services must store the notification with correctly info$`:  s.NotificationMgmtMustStoreTheNotification,
		`^update correctly corresponding field$`:                                       s.updateCorrectlyCorrespondingField,
		`^"([^"]*)" update the notification with location filter change to "([^"]*)"$`: s.StaffUpdateNotificationWithLocationFilterChange,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *AccessControlForUpdateNotificationSuite) updateCorrectlyCorrespondingField(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	resp, ok := commonState.Response.(*npb.UpsertNotificationResponse)
	if !ok {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("expect npb.UpsertNotificationResponse but got %v", commonState.Response)
	}

	if commonState.Notification.NotificationId != resp.NotificationId {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("expect upsert with the same notifition id %v but got %v", commonState.Notification.NotificationId, resp.NotificationId)
	}

	return s.NotificationMgmtMustStoreTheNotification(ctx)
}

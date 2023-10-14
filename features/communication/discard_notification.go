package communication

import (
	"github.com/manabie-com/backend/features/communication/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
)

type DiscardNotificationSuite struct {
	*common.NotificationSuite
	NotificationNeedToSent *cpb.Notification
}

func (c *SuiteConstructor) InitDiscardNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &DiscardNotificationSuite{
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
		`^current staff discards notification$`:                      s.CurrentStaffDiscardsNotification,
		`^notification is discarded$`:                                s.NotificationIsDiscarded,
		`^returns "([^"]*)" status code$`:                            s.CheckReturnStatusCode,
		`^returns error message "([^"]*)"$`:                          s.CheckReturnsErrorMessage,
		`^notificationmgmt services must send notification to user$`: s.NotificationMgmtMustSendNotificationToUser,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

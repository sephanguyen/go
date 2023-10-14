package communication

import (
	"github.com/manabie-com/backend/features/communication/common"

	"github.com/cucumber/godog"
)

type AccessControlForCreateNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitAccessControlForCreateNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &AccessControlForCreateNotificationSuite{
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
		`^returns "([^"]*)" status code$`:                                             s.CheckReturnStatusCode,
		`^notificationmgmt services must store the notification with correctly info$`: s.NotificationMgmtMustStoreTheNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

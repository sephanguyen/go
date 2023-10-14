package communication

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4"
)

type AccessControlForGetNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitAccessControlForGetNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &AccessControlForGetNotificationSuite{
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
		`^notificationmgmt services must store the notification with correctly info$`:  s.NotificationMgmtMustStoreTheNotification,
		`^update correctly corresponding field$`:                                       s.updateCorrectlyCorrespondingField,
		`^"([^"]*)" staff get the created notification "([^"]*)"$`:                     s.staffGetTheCreatedNotification,
		`^"([^"]*)" update the notification with location filter change to "([^"]*)"$`: s.StaffUpdateNotificationWithLocationFilterChange,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *AccessControlForGetNotificationSuite) updateCorrectlyCorrespondingField(ctx context.Context) (context.Context, error) {
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

func (s *AccessControlForGetNotificationSuite) staffGetTheCreatedNotification(ctx context.Context, typeStaff string, statusGet string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	var ctxWithUserID context.Context
	if typeStaff == "current" {
		ctxWithUserID = contextWithUserID(ctx, commonState.CurrentUserID)
	} else {
		ctxWithUserID = contextWithUserID(ctx, commonState.Organization.Staffs[len(commonState.Organization.Staffs)-2].ID)
	}

	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctxWithUserID, s.BobDBConn, queryGetNotication, commonState.Notification.NotificationId).ScanOne(infoNotification)

	if statusGet == "successfully" && err != nil {
		return ctx, fmt.Errorf("expected no error, got: %v", err)
	}

	if statusGet == "failed" && err != nil && err != pgx.ErrNoRows {
		return ctx, fmt.Errorf("expected no rows error, got: %v", err)
	}

	return ctx, nil
}

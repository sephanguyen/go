package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/features/communication/common/helpers"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"k8s.io/utils/strings/slices"
)

type CreateAndUpdateNotificationWithAcessPathSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitCreateAndUpdateNotificationWithAccessPath(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &CreateAndUpdateNotificationWithAcessPathSuite{
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
		`^returns "([^"]*)" status code$`:                                                  s.CheckReturnStatusCode,
		`^notificationmgmt services must store the notification with correctly info$`:      s.NotificationMgmtMustStoreTheNotification,
		`^the notification access path has been store correctly with "([^"]*)" locations$`: s.checkNotificationAccessPath,
		`^"([^"]*)" update the notification with location filter change to "([^"]*)"$`:     s.StaffUpdateNotificationWithLocationFilterChange,
		`^update correctly corresponding field$`:                                           s.updateCorrectlyCorrespondingField,
		`^admin update staff granted locations to "([^"]*)"$`:                              s.AdminUpdateCurrentStaffGrantedLocationsTo,
		`^current staff upsert with "([^"]*)" locations and send notification$`:            s.currentStaffUpsertAndSendNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateAndUpdateNotificationWithAcessPathSuite) checkNotificationAccessPath(ctx context.Context, idxLocStr string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	expectedLocationIDs := make([]string, 0)
	idxsLocsStr := strings.Split(idxLocStr, ",")
	for _, idxLocStr := range idxsLocsStr {
		if idxLocStr == "default" {
			expectedLocationIDs = append(expectedLocationIDs, commonState.Organization.DefaultLocation.ID)
			continue
		}

		idxLoc, err := strconv.Atoi(idxLocStr)
		if err != nil {
			return ctx, fmt.Errorf("can't convert descendant location index: %v", err)
		}
		if idxLoc <= 0 || idxLoc > helpers.NumberOfNewCenterLocationCreated {
			return ctx, fmt.Errorf("index descendant location out of range")
		}
		expectedLocationIDs = append(expectedLocationIDs, commonState.Organization.DescendantLocations[idxLoc-1].ID)
	}

	queryGetNoticationAccessPath := `
		SELECT location_id 
		FROM info_notifications_access_paths 
		WHERE notification_id = $1 AND deleted_at IS NULL;
	`
	rows, err := s.BobDBConn.Query(ctx, queryGetNoticationAccessPath, commonState.Notification.NotificationId)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}
	defer rows.Close()

	actualLocationIDs := make([]string, 0)
	for rows.Next() {
		var locationID pgtype.Text
		err = rows.Scan(&locationID)
		if err != nil {
			return common.StepStateToContext(ctx, commonState), err
		}

		actualLocationIDs = append(actualLocationIDs, locationID.String)
		if !slices.Contains(expectedLocationIDs, locationID.String) {
			return ctx, fmt.Errorf("unexpect location %s in (%+v)", locationID.String, expectedLocationIDs)
		}
	}

	if !stringutil.SliceElementsMatch(expectedLocationIDs, actualLocationIDs) {
		return ctx, fmt.Errorf("expect elements of location IDs %+v, got %+v", expectedLocationIDs, actualLocationIDs)
	}

	if diffArr := stringutil.SliceElementsDiff(actualLocationIDs, expectedLocationIDs); len(diffArr) > 0 {
		return ctx, fmt.Errorf("unexpect locations %+v", diffArr)
	}

	return ctx, nil
}

func (s *CreateAndUpdateNotificationWithAcessPathSuite) updateCorrectlyCorrespondingField(ctx context.Context) (context.Context, error) {
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

func (s *CreateAndUpdateNotificationWithAcessPathSuite) currentStaffUpsertAndSendNotification(ctx context.Context, locationFilter string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	notification := commonState.Notification

	locationIDs, selectType, err := s.GetLocationIDsFromString(ctx, locationFilter)
	if err != nil {
		return ctx, err
	}
	notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
		Type:        selectType,
		LocationIds: locationIDs,
	}

	req := &npb.UpsertNotificationRequest{
		Notification: notification,
	}
	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
		s.ContextWithToken(ctx, commonState.CurrentStaff.Token),
		req)
	if err != nil {
		return ctx, fmt.Errorf("failed UpsertNotification: %v", err)
	}
	return s.CurrentStaffSendNotification(common.StepStateToContext(ctx, commonState))
}

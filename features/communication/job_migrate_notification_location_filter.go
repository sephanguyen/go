package communication

import (
	"context"
	"fmt"
	"strings"

	notificatio_cmd "github.com/manabie-com/backend/cmd/server/notificationmgmt"
	"github.com/manabie-com/backend/features/communication/common"

	"github.com/cucumber/godog"
	"go.uber.org/zap"
)

type JobMigrateNotificationLocationFilterSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitJobMigrateNotificationLocationFilter(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JobMigrateNotificationLocationFilterSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`:                                                                                                                                  s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^school admin creates "([^"]*)" courses$`:     s.CreatesNumberOfCourses,
		`^run migration script$`:                       s.runMigrationScript,
		`^data of target group is correctly migrated$`: s.dataOfTargetGroupIsCorrectlyMigrated,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JobMigrateNotificationLocationFilterSuite) runMigrationScript(ctx context.Context) (context.Context, error) {
	err := notificatio_cmd.MigrateNotificationLocationFilter(ctx, notiConfig, s.BobDBConn, zap.NewNop())
	if err != nil {
		return ctx, fmt.Errorf("err migrate: %v", err)
	}

	return ctx, nil
}

func (s *JobMigrateNotificationLocationFilterSuite) dataOfTargetGroupIsCorrectlyMigrated(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	// find notification that has SELECT LIST but has not been migrated
	query := `
		SELECT in2.notification_id
		FROM info_notifications in2
		LEFT JOIN notification_location_filter nlf ON nlf.notification_id = in2.notification_id 
		WHERE in2.target_groups ->'location_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
		AND in2.resource_path = $1
		AND nlf.notification_id IS NULL;
	`

	rows, err := s.BobDBConn.Query(ctx, query, commonState.CurrentResourcePath)
	if err != nil {
		return ctx, fmt.Errorf("failed query: %v", err)
	}

	defer rows.Close()
	notificationIDs := []string{}
	for rows.Next() {
		var notificationID string
		err = rows.Scan(&notificationID)
		if err != nil {
			return ctx, fmt.Errorf("failed scan: %v", err)
		}
		notificationIDs = append(notificationIDs, notificationID)
	}
	if len(notificationIDs) > 0 {
		return ctx, fmt.Errorf("notification ID %s has not migrated", strings.Join(notificationIDs, ","))
	}

	// find notification has been migrated but wrong location_id
	query = `
	SELECT nlf.notification_id
	FROM notification_location_filter nlf 
	LEFT JOIN (
		SELECT in2.notification_id,
		jsonb_array_elements_text(in2.target_groups -> 'location_filter'->'location_ids') AS location_id
		FROM info_notifications in2 
		WHERE in2.target_groups ->'location_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
	) tmp1 ON tmp1.notification_id = nlf.notification_id AND tmp1.location_id = nlf.location_id
	WHERE tmp1.notification_id IS NULL AND nlf.resource_path=$1;
	`
	rows, err = s.BobDBConn.Query(ctx, query, commonState.CurrentResourcePath)
	if err != nil {
		return ctx, fmt.Errorf("failed query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var notificationID string
		err = rows.Scan(&notificationID)
		if err != nil {
			return ctx, fmt.Errorf("failed scan: %v", err)
		}
		notificationIDs = append(notificationIDs, notificationID)
	}
	if len(notificationIDs) > 0 {
		return ctx, fmt.Errorf("notifications %s has migrated with wrong course data", strings.Join(notificationIDs, ","))
	}
	return ctx, nil
}

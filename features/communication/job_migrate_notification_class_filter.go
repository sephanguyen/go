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

type JobMigrateNotificationClassFilterSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitJobMigrateNotificationClassFilter(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JobMigrateNotificationClassFilterSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`:                                                                                                                                  s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^run migration script$`:                                                          s.runMigrationScript,
		`^data of target group is correctly migrated$`:                                    s.dataOfTargetGroupIsCorrectlyMigrated,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`: s.CreatesNumberOfCoursesWithClass,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JobMigrateNotificationClassFilterSuite) runMigrationScript(ctx context.Context) (context.Context, error) {
	err := notificatio_cmd.MigrateNotificationClassFilter(ctx, notiConfig, s.BobDBConn, zap.NewNop())
	if err != nil {
		return ctx, fmt.Errorf("err migrate: %v", err)
	}

	return ctx, nil
}

func (s *JobMigrateNotificationClassFilterSuite) dataOfTargetGroupIsCorrectlyMigrated(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	// find notification that has SELECT LIST but has not been migrated
	query := `
		SELECT in2.notification_id
		FROM info_notifications in2
		LEFT JOIN notification_class_filter ncf ON ncf.notification_id = in2.notification_id 
		WHERE in2.target_groups ->'class_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
		AND in2.resource_path = $1
		AND ncf.notification_id IS NULL;
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

	// find notification has been migrated but wrong class_id
	query = `
	SELECT ncf.notification_id
	FROM notification_class_filter ncf 
	LEFT JOIN (
		SELECT in2.notification_id,
		jsonb_array_elements_text(in2.target_groups -> 'class_filter'->'class_ids') AS class_id
		FROM info_notifications in2 
		WHERE in2.target_groups ->'class_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
	) tmp1 ON tmp1.notification_id = ncf.notification_id AND tmp1.class_id = ncf.class_id
	WHERE tmp1.notification_id IS NULL AND ncf.resource_path=$1;
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
		return ctx, fmt.Errorf("notifications %s has migrated with wrong class data", strings.Join(notificationIDs, ","))
	}
	return ctx, nil
}

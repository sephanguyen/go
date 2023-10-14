package notificationmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/notification/config"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("notificationmgmt_migrate_notification_course_filter", RunMigrateNotificationCourseFilter)
}

func RunMigrateNotificationCourseFilter(ctx context.Context, notiCfg config.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("bob")
	zLogger := rsc.Logger()

	return MigrateNotificationCourseFilter(ctx, notiCfg, db.DB.(*pgxpool.Pool), zLogger)
}

func MigrateNotificationCourseFilter(ctx context.Context, notiCfg config.Config, dbPool *pgxpool.Pool, zLogger *zap.Logger) error {
	if zLogger == nil {
		zLogger = logger.NewZapLogger("debug", notiCfg.Common.Environment == localEnv)
	}

	organizations, err := dbPool.Query(ctx, scanOrganiationQuery)
	if err != nil {
		return fmt.Errorf("get orgs failed: %s", err)
	}
	defer organizations.Close()

	// Migrate with RP
	totalNotiProcessed := 0
	for organizations.Next() {
		var organizationID string
		err := organizations.Scan(&organizationID)
		if err != nil {
			return fmt.Errorf("failed to scan an orgs row: %w", err)
		}

		tenantAndUserCtx, err := makeTenantWithUserCtx(ctx, dbPool, organizationID)
		if err != nil {
			zLogger.Sugar().Error(err)
		}

		totalNotiProcessedPerTenant, err := migrateNotificationCourseFilter(tenantAndUserCtx, dbPool, organizationID)
		if err != nil {
			return fmt.Errorf("migrateNotificationCourseName: %s", err)
		}
		totalNotiProcessed += totalNotiProcessedPerTenant
	}
	zLogger.Sugar().Infof("There is/are %d notification migrated", totalNotiProcessed)
	zLogger.Sugar().Info("----- DONE: Migrating notification course filter job -----")
	return nil
}

func migrateNotificationCourseFilter(ctx context.Context, dbPool *pgxpool.Pool, organizationID string) (int, error) {
	totalProcessedPerTenant := 0
	// only select notification that has course_ids to extract and migrate
	conditionStr := `
	AND in2.target_groups -> 'course_filter' IS NOT NULL
	AND in2.target_groups->'course_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
	AND in2.target_groups->'course_filter'->'course_ids' != '[]'::jsonb
	`
	if err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		offset := 0
		notifications, err := getInfoNotificationWithOffset(ctx, tx, offset, organizationID, conditionStr)
		if err != nil {
			return fmt.Errorf("failed getInfoNotificationWithOffset: %s", err)
		}

		mapNotificationIDAndCourseIDs := make(map[string][]string, 0)
		for len(notifications) > 0 {
			for _, noti := range notifications {
				targetGroup := &entities.InfoNotificationTarget{}
				err = noti.TargetGroups.AssignTo(targetGroup)
				if err != nil {
					return fmt.Errorf("failed assign target group: %v", err)
				}

				if len(targetGroup.CourseFilter.CourseIDs) > 0 {
					mapNotificationIDAndCourseIDs[noti.NotificationID.String] = targetGroup.CourseFilter.CourseIDs
				}
			}

			err = BulkInsertNotificationFilter(ctx, tx, "course", organizationID, mapNotificationIDAndCourseIDs)
			if err != nil {
				return fmt.Errorf("failed BulkInsertNotificationFilter: %v", err)
			}
			totalProcessedPerTenant += len(notifications)

			offset += len(notifications)
			notifications, err = getInfoNotificationWithOffset(ctx, tx, offset, organizationID, conditionStr)
			if err != nil {
				return fmt.Errorf("failed getInfoNotificationWithOffset")
			}
		}
		return nil
	}); err != nil {
		return 0, fmt.Errorf("database.ExecInTx err: %v", err)
	}

	return totalProcessedPerTenant, nil
}

func BulkInsertNotificationFilter(ctx context.Context, db database.QueryExecer, field, organizationID string, mapNotificationIDAndValues map[string][]string) error {
	now := database.Timestamptz(time.Now())
	queueFn := func(b *pgx.Batch, notificationID, fieldID string) {
		query := `
			INSERT INTO notification_%s_filter (notification_id, %s_id, created_at, updated_at, resource_path)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (notification_id, %s_id) DO NOTHING;
		`

		b.Queue(fmt.Sprintf(query, field, field, field),
			database.Text(notificationID),
			database.Text(fieldID),
			now,
			now,
			database.Text(organizationID))
	}

	b := &pgx.Batch{}
	totalExecutes := 0
	for notiID, values := range mapNotificationIDAndValues {
		for _, val := range values {
			queueFn(b, notiID, val)
		}
		totalExecutes += len(values)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i := 0; i < totalExecutes; i++ {
		if _, err := result.Exec(); err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

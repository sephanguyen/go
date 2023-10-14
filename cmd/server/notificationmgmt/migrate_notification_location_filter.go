package notificationmgmt

import (
	"context"
	"fmt"

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
	bootstrap.RegisterJob("notificationmgmt_migrate_notification_location_filter", RunMigrateNotificationLocationFilter)
}

func RunMigrateNotificationLocationFilter(ctx context.Context, notiCfg config.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("bob")
	zLogger := rsc.Logger()

	return MigrateNotificationLocationFilter(ctx, notiCfg, db.DB.(*pgxpool.Pool), zLogger)
}

func MigrateNotificationLocationFilter(ctx context.Context, notiCfg config.Config, dbPool *pgxpool.Pool, zLogger *zap.Logger) error {
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

		totalNotiProcessedPerTenant, err := migrateNotificationLocationFilter(tenantAndUserCtx, dbPool, organizationID)
		if err != nil {
			return fmt.Errorf("migrateNotificationLocationFilter: %s", err)
		}
		totalNotiProcessed += totalNotiProcessedPerTenant
	}
	zLogger.Sugar().Infof("There is/are %d notification migrated", totalNotiProcessed)
	zLogger.Sugar().Info("----- DONE: Migrating notification location filter job -----")
	return nil
}

func migrateNotificationLocationFilter(ctx context.Context, dbPool *pgxpool.Pool, organizationID string) (int, error) {
	totalProcessedPerTenant := 0
	// only select notification that has location_ids to extract and migrate
	conditionStr := `
	AND in2.target_groups -> 'location_filter' IS NOT NULL
	AND in2.target_groups->'location_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_LIST'
	AND in2.target_groups->'location_filter'->'location_ids' != '[]'::jsonb
	`
	if err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		offset := 0
		notifications, err := getInfoNotificationWithOffset(ctx, tx, offset, organizationID, conditionStr)
		if err != nil {
			return fmt.Errorf("failed getInfoNotificationWithOffset: %s", err)
		}

		mapNotificationIDAndLocationIDs := make(map[string][]string, 0)
		for len(notifications) > 0 {
			for _, noti := range notifications {
				targetGroup := &entities.InfoNotificationTarget{}
				err = noti.TargetGroups.AssignTo(targetGroup)
				if err != nil {
					return fmt.Errorf("failed assign target group: %v", err)
				}

				if len(targetGroup.LocationFilter.LocationIDs) > 0 {
					mapNotificationIDAndLocationIDs[noti.NotificationID.String] = targetGroup.LocationFilter.LocationIDs
				}
			}

			err = BulkInsertNotificationFilter(ctx, tx, "location", organizationID, mapNotificationIDAndLocationIDs)
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

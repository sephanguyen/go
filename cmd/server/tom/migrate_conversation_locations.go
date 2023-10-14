package tom

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	config "github.com/manabie-com/backend/internal/tom/configurations"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("tom_migrate_conversation_locations", RunMigrateConversationLocations)
}

func RunMigrateConversationLocations(ctx context.Context, tomCfg config.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	zLogger := rsc.Logger()

	return MigrateConversationLocations(ctx, tomCfg, db, zLogger)
}

func MigrateConversationLocations(ctx context.Context, tomCfg config.Config, tomDBTrace *database.DBTrace, zLogger *zap.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if zLogger == nil {
		zLogger = logger.NewZapLogger("debug", tomCfg.Common.Environment == "local")
	}

	organizations, err := tomDBTrace.Query(ctx,
		`SELECT organization_id, name 
		FROM organizations`)
	if err != nil {
		return fmt.Errorf("get orgs failed")
	}
	defer organizations.Close()

	for organizations.Next() {
		var organizationID, name string
		err := organizations.Scan(&organizationID, &name)
		if err != nil {
			zLogger.Sugar().Errorf("failed to scan an orgs row: %w", err)
			continue
		}
		ctx = auth.InjectFakeJwtToken(ctx, organizationID)

		if organizationID == "" {
			zLogger.Sugar().Fatal("running is requires a school id")
			continue
		}

		rows, err := tomDBTrace.Query(ctx, `
			SELECT cs.conversation_id as conversation_id
				, ap.location_id as location_id
				, ap.location_id as access_path
			FROM conversation_students cs
			LEFT JOIN conversation_locations cl ON cs.conversation_id = cl.conversation_id
			INNER JOIN user_access_paths ap ON cs.student_id = ap.user_id
			INNER JOIN locations lc ON ap.location_id = lc.location_id
			INNER JOIN location_types lt ON lc.location_type = lt.location_type_id
			WHERE cl.conversation_id IS NULL
			AND cs.deleted_at IS NULL
			AND cl.deleted_at IS NULL
			AND ap.deleted_at IS NULL
			AND lc.deleted_at IS NULL
			AND lt.deleted_at IS NULL
			AND lt."name" = 'org'
			AND cs.resource_path = $1
			`, organizationID)
		if err != nil {
			zLogger.Sugar().Fatalf("Error at querying conversation locations: %w", err)
		}
		defer rows.Close()

		conversationLocations := []domain.ConversationLocation{}
		for rows.Next() {
			var conversationID, locationID, accessPath pgtype.Text
			if err := rows.Scan(&conversationID, &locationID, &accessPath); err != nil {
				zLogger.Sugar().Fatalf("error scan row: %v", err)
			}

			now := time.Now()

			var e domain.ConversationLocation
			database.AllNullEntity(&e)
			err := multierr.Combine(
				e.ConversationID.Set(conversationID),
				e.LocationID.Set(locationID),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			)
			if err != nil {
				zLogger.Sugar().Fatalf("error set conversation location: %w", err)
			}
			conversationLocations = append(conversationLocations, e)
		}

		ConversationLocationRepo := &repositories.ConversationLocationRepo{}

		if err := ConversationLocationRepo.BulkUpsert(ctx, tomDBTrace, conversationLocations); err != nil {
			zLogger.Sugar().Errorf("ConversationLocationRepo.BulkCreate: %w", err)
		}
	}

	return nil
}

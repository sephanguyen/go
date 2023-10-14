package repo

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type ActivityLogRepo struct{}

func (r *ActivityLogRepo) Create(ctx context.Context, db database.Ext, userID, actionType string, payload map[string]interface{}) error {
	ctx, span := interceptors.StartSpan(ctx, "ActivityLogRepo.Create")
	defer span.End()

	log := &bob_entities.ActivityLog{}
	now := time.Now()
	err := multierr.Combine(
		log.UserID.Set(userID),
		log.ActionType.Set(actionType),
		log.Payload.Set(payload),
		log.CreatedAt.Set(now),
		log.UpdatedAt.Set(now),
		log.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	if log.ID.String == "" {
		err = log.ID.Set(idutil.ULIDNow())
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}
	}

	cmdTag, err := database.Insert(ctx, log, db.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot insert new ActivityLog")
	}

	return nil
}

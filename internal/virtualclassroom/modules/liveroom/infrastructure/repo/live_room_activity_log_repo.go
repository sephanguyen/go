package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"

	"go.uber.org/multierr"
)

type LiveRoomActivityLogRepo struct{}

func (l *LiveRoomActivityLogRepo) CreateLog(ctx context.Context, db database.Ext, channelID, userID, actionType string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomActivityLogRepo.CreateLog")
	defer span.End()

	log := &LiveRoomActivityLog{}
	now := time.Now()
	if err := multierr.Combine(
		log.ActivityLogID.Set(idutil.ULIDNow()),
		log.ChannelID.Set(channelID),
		log.UserID.Set(userID),
		log.ActionType.Set(actionType),
		log.CreatedAt.Set(now),
		log.UpdatedAt.Set(now),
		log.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	cmdTag, err := database.Insert(ctx, log, db.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return domain.ErrNoLiveRoomActivityLogCreated
	}

	return nil
}

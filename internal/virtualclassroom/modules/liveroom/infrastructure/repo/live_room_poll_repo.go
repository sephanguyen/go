package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
)

type LiveRoomPollRepo struct{}

func (l *LiveRoomPollRepo) CreateLiveRoomPoll(ctx context.Context, db database.QueryExecer, liveRoomPoll *domain.LiveRoomPoll) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomPollRepo.CreateLiveRoomPoll")
	defer span.End()

	liveRoomPollDTO, err := NewLiveRoomPollFromEntity(liveRoomPoll)
	if err != nil {
		return err
	}

	if err := liveRoomPollDTO.PreInsert(); err != nil {
		return err
	}

	fields := database.GetFieldNamesExcepts(liveRoomPollDTO, []string{"deleted_at"})
	placeHolders := database.GeneratePlaceholders(len(fields))
	args := database.GetScanFields(liveRoomPollDTO, fields)

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		liveRoomPollDTO.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

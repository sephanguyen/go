package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/pkg/errors"
)

type LiveRoomLogRepo struct{}

const getLatestLogIDByChannelIDQuery = `SELECT live_room_log_id FROM live_room_log 
	WHERE channel_id = $1 AND deleted_at is null AND is_completed = FALSE 
	ORDER BY created_at DESC LIMIT 1`

func (l *LiveRoomLogRepo) Create(ctx context.Context, db database.QueryExecer, dto *LiveRoomLog) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomLogRepo.Create")
	defer span.End()

	if err := dto.PreInsert(); err != nil {
		return err
	}

	cmdTag, err := database.Insert(ctx, dto, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new live_room_log")
	}

	return nil
}

func (l *LiveRoomLogRepo) AddAttendeeIDByChannelID(ctx context.Context, db database.QueryExecer, channelID, attendeeID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomLogRepo.AddAttendeeIDByChannelID")
	defer span.End()

	query := fmt.Sprintf(`UPDATE live_room_log 
		SET attendee_ids = array_append(attendee_ids, $2), updated_at = now() 
		WHERE NOT($2 = ANY(attendee_ids)) AND live_room_log_id IN (%s)`,
		getLatestLogIDByChannelIDQuery,
	)
	_, err := db.Exec(ctx, query, channelID, attendeeID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LiveRoomLogRepo) IncreaseTotalTimesByChannelID(ctx context.Context, db database.QueryExecer, channelID string, logType TotalTimes) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomLogRepo.IncreaseTotalTimesByChannelID")
	defer span.End()

	var willUpdateField string
	switch logType {
	case TotalTimesReconnection:
		willUpdateField = "total_times_reconnection"
	case TotalTimesUpdatingRoomState:
		willUpdateField = "total_times_updating_room_state"
	case TotalTimesGettingRoomState:
		willUpdateField = "total_times_getting_room_state"
	default:
		return fmt.Errorf("live room log type unsupported %v", logType)
	}

	query := fmt.Sprintf(`UPDATE live_room_log 
		SET :willUpdateField = coalesce(:willUpdateField, 0) + 1, updated_at = now() 
		WHERE live_room_log_id IN (%s)`,
		getLatestLogIDByChannelIDQuery,
	)
	query = strings.ReplaceAll(query, ":willUpdateField", willUpdateField)

	_, err := db.Exec(ctx, query, channelID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LiveRoomLogRepo) CompleteLogByChannelID(ctx context.Context, db database.QueryExecer, channelID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomLogRepo.CompleteLogByChannelID")
	defer span.End()

	query := fmt.Sprintf(`UPDATE live_room_log 
		SET is_completed = TRUE, updated_at = now() 
		WHERE live_room_log_id IN (%s)`,
		getLatestLogIDByChannelIDQuery,
	)

	_, err := db.Exec(ctx, query, channelID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LiveRoomLogRepo) GetLatestByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*LiveRoomLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomLogRepo.GetLatestByChannelID")
	defer span.End()

	dto := &LiveRoomLog{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
			WHERE channel_id = $1 
			AND deleted_at is null 
			ORDER BY created_at DESC LIMIT 1`,
		strings.Join(fields, ","),
		dto.TableName(),
	)

	err := db.QueryRow(ctx, query, &channelID).Scan(values...)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

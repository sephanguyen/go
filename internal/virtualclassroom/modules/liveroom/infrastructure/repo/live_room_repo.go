package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LiveRoomRepo struct{}

func (l *LiveRoomRepo) GetLiveRoomByChannelName(ctx context.Context, db database.QueryExecer, channelName string) (*domain.LiveRoom, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRepo.GetLiveRoomByChannelName")
	defer span.End()

	liveRoom := &LiveRoom{}
	fields, values := liveRoom.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s
			WHERE channel_name = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		liveRoom.TableName(),
	)
	err := db.QueryRow(ctx, query, &channelName).Scan(values...)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrChannelNotFound
	} else if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	res := domain.NewLiveRoom().
		WithChannelID(liveRoom.ChannelID.String).
		WithChannelName(liveRoom.ChannelName.String).
		WithWhiteboardRoomID(liveRoom.WhiteboardRoomID.String).
		WithEndedAt(database.FromTimestamptz(liveRoom.EndedAt)).
		WithModifiedTime(liveRoom.CreatedAt.Time, liveRoom.UpdatedAt.Time).
		WithDeletedAt(database.FromTimestamptz(liveRoom.DeletedAt)).
		BuildDraft()

	return res, nil
}

func (l *LiveRoomRepo) CreateLiveRoom(ctx context.Context, db database.QueryExecer, channelID, channelName, roomUUID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRepo.CreateLiveRoom")
	defer span.End()

	liveRoomDTO, err := NewLiveRoomFromEntity(&domain.LiveRoom{
		ChannelID:        channelID,
		ChannelName:      channelName,
		WhiteboardRoomID: roomUUID,
	})
	if err != nil {
		return err
	}

	if err := liveRoomDTO.PreInsert(); err != nil {
		return err
	}

	fields := database.GetFieldNamesExcepts(liveRoomDTO, []string{"ended_at", "deleted_at"})
	placeHolders := database.GeneratePlaceholders(len(fields))
	args := database.GetScanFields(liveRoomDTO, fields)

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT unique__channel_name DO NOTHING`,
		liveRoomDTO.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)
	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return domain.ErrNoChannelCreated
	}

	return nil
}

func (l *LiveRoomRepo) GetLiveRoomByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*domain.LiveRoom, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRepo.GetLiveRoomByChannelID")
	defer span.End()

	liveRoom := &LiveRoom{}
	fields, values := liveRoom.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s
			WHERE channel_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		liveRoom.TableName(),
	)
	err := db.QueryRow(ctx, query, &channelID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	res := domain.NewLiveRoom().
		WithChannelID(liveRoom.ChannelID.String).
		WithChannelName(liveRoom.ChannelName.String).
		WithWhiteboardRoomID(liveRoom.WhiteboardRoomID.String).
		WithEndedAt(database.FromTimestamptz(liveRoom.EndedAt)).
		WithModifiedTime(liveRoom.CreatedAt.Time, liveRoom.UpdatedAt.Time).
		WithDeletedAt(database.FromTimestamptz(liveRoom.DeletedAt)).
		BuildDraft()

	return res, nil
}

func (l *LiveRoomRepo) EndLiveRoom(ctx context.Context, db database.QueryExecer, channelID string, endTime time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRepo.EndLiveRoom")
	defer span.End()

	var endedAt pgtype.Timestamptz
	if err := endedAt.Set(endTime); err != nil {
		return fmt.Errorf("endedAt.Set: %w", err)
	}

	query := `UPDATE live_room 
		SET ended_at = $1
		WHERE channel_id = $2`

	cmdTag, err := db.Exec(ctx, query, &endedAt, &channelID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("channel %s end time was not updated", channelID)
	}

	return nil
}

func (l *LiveRoomRepo) UpdateChannelRoomID(ctx context.Context, db database.QueryExecer, channelID, roomUUID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRepo.UpdateChannelRoomID")
	defer span.End()

	query := `UPDATE live_room 
		SET updated_at = now(), whiteboard_room_id = $1
		WHERE channel_id = $2`

	cmdTag, err := db.Exec(ctx, query, &roomUUID, &channelID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("cannot update channel %s room ID %s", channelID, roomUUID)
	}

	return nil
}

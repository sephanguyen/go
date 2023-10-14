package repo

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveRoomLog struct {
	LiveRoomLogID               pgtype.Text
	ChannelID                   pgtype.Text
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	DeletedAt                   pgtype.Timestamptz
	IsCompleted                 pgtype.Bool
	AttendeeIDs                 pgtype.TextArray
	TotalTimesReconnection      pgtype.Int4
	TotalTimesUpdatingRoomState pgtype.Int4
	TotalTimesGettingRoomState  pgtype.Int4
}

func (l *LiveRoomLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"live_room_log_id",
			"channel_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"is_completed",
			"attendee_ids",
			"total_times_reconnection",
			"total_times_updating_room_state",
			"total_times_getting_room_state",
		}, []interface{}{
			&l.LiveRoomLogID,
			&l.ChannelID,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
			&l.IsCompleted,
			&l.AttendeeIDs,
			&l.TotalTimesReconnection,
			&l.TotalTimesUpdatingRoomState,
			&l.TotalTimesGettingRoomState,
		}
}

func (l *LiveRoomLog) TableName() string {
	return "live_room_log"
}

func (l *LiveRoomLog) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("failed to set values in PreInsert: %w", err)
	}

	return nil
}

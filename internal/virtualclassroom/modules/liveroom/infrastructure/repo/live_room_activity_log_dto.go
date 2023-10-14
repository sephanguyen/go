package repo

import "github.com/jackc/pgtype"

type LiveRoomActivityLog struct {
	ActivityLogID pgtype.Text
	ChannelID     pgtype.Text
	UserID        pgtype.Text
	ActionType    pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
}

func (l *LiveRoomActivityLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"activity_log_id",
			"channel_id",
			"user_id",
			"action_type",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.ActivityLogID,
			&l.ChannelID,
			&l.UserID,
			&l.ActionType,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LiveRoomActivityLog) TableName() string {
	return "live_room_activity_logs"
}

package repo

import "github.com/jackc/pgtype"

type TotalTimes int

const (
	TotalTimesReconnection TotalTimes = iota
	TotalTimesUpdatingRoomState
	TotalTimesGettingRoomState
)

type VirtualClassRoomLogDTO struct {
	LogID                       pgtype.Text
	LessonID                    pgtype.Text
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	DeletedAt                   pgtype.Timestamptz
	IsCompleted                 pgtype.Bool
	AttendeeIDs                 pgtype.TextArray
	TotalTimesReconnection      pgtype.Int4
	TotalTimesUpdatingRoomState pgtype.Int4
	TotalTimesGettingRoomState  pgtype.Int4
}

func (v *VirtualClassRoomLogDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"log_id",
			"lesson_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"is_completed",
			"attendee_ids",
			"total_times_reconnection",
			"total_times_updating_room_state",
			"total_times_getting_room_state",
		}, []interface{}{
			&v.LogID,
			&v.LessonID,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.DeletedAt,
			&v.IsCompleted,
			&v.AttendeeIDs,
			&v.TotalTimesReconnection,
			&v.TotalTimesUpdatingRoomState,
			&v.TotalTimesGettingRoomState,
		}
}

func (v *VirtualClassRoomLogDTO) TableName() string {
	return "virtual_classroom_log"
}

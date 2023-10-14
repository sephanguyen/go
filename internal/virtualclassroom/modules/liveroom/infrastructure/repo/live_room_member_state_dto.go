package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveRoomMemberState struct {
	ChannelID        pgtype.Text
	UserID           pgtype.Text
	StateType        pgtype.Text
	BoolValue        pgtype.Bool
	StringArrayValue pgtype.TextArray
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (l *LiveRoomMemberState) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"channel_id",
		"user_id",
		"state_type",
		"bool_value",
		"string_array_value",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&l.ChannelID,
		&l.UserID,
		&l.StateType,
		&l.BoolValue,
		&l.StringArrayValue,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.DeletedAt,
	}
	return
}

func (l *LiveRoomMemberState) TableName() string {
	return "live_room_member_state"
}

func (l *LiveRoomMemberState) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("failed to set values in PreInsert: %w", err)
	}

	return nil
}

type LiveRoomMemberStates []*LiveRoomMemberState

func (ls *LiveRoomMemberStates) Add() database.Entity {
	e := &LiveRoomMemberState{}
	*ls = append(*ls, e)

	return e
}

func (ls LiveRoomMemberStates) ToLiveRoomMemberStatesDomain() domain.LiveRoomMemberStates {
	lms := make(domain.LiveRoomMemberStates, 0, len(ls))

	for _, states := range ls {
		strArrayValue := make([]string, 0, len(states.StringArrayValue.Elements))
		for _, str := range states.StringArrayValue.Elements {
			strArrayValue = append(strArrayValue, str.String)
		}

		lmsDomain := &domain.LiveRoomMemberState{
			ChannelID:        states.ChannelID.String,
			UserID:           states.UserID.String,
			StateType:        states.StateType.String,
			BoolValue:        states.BoolValue.Bool,
			StringArrayValue: strArrayValue,
			CreatedAt:        states.CreatedAt.Time,
			UpdatedAt:        states.UpdatedAt.Time,
		}
		if states.DeletedAt.Status == pgtype.Present {
			lmsDomain.DeletedAt = &states.DeletedAt.Time
		}

		lms = append(lms, lmsDomain)
	}

	return lms
}

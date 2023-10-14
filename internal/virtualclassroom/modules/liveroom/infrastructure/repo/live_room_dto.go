package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveRoom struct {
	ChannelID        pgtype.Text
	ChannelName      pgtype.Text
	WhiteboardRoomID pgtype.Text
	EndedAt          pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func NewLiveRoomFromEntity(l *domain.LiveRoom) (*LiveRoom, error) {
	dto := &LiveRoom{}
	database.AllNullEntity(dto)

	if err := multierr.Combine(
		dto.ChannelID.Set(l.ChannelID),
		dto.ChannelName.Set(l.ChannelName),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not map live room entity to dto: %w", err)
	}

	if len(l.WhiteboardRoomID) > 0 {
		if err := dto.WhiteboardRoomID.Set(l.WhiteboardRoomID); err != nil {
			return nil, fmt.Errorf("could not map live room entity to dto, whiteboard room id: %w", err)
		}
	}

	if l.EndedAt != nil {
		if err := dto.EndedAt.Set(l.EndedAt); err != nil {
			return nil, fmt.Errorf("could not map live room entity to dto, ended at: %w", err)
		}
	}

	if l.DeletedAt != nil {
		if err := dto.DeletedAt.Set(l.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not map live room entity to dto, deleted at: %w", err)
		}
	}

	return dto, nil
}

func (l *LiveRoom) FieldMap() ([]string, []interface{}) {
	return []string{
			"channel_id",
			"channel_name",
			"whiteboard_room_id",
			"ended_at",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.ChannelID,
			&l.ChannelName,
			&l.WhiteboardRoomID,
			&l.EndedAt,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LiveRoom) TableName() string {
	return "live_room"
}

func (l *LiveRoom) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("failed to set values in PreInsert: %w", err)
	}

	return nil
}

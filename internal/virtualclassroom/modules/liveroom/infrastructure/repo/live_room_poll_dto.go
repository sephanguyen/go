package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveRoomPoll struct {
	LiveRoomPollID pgtype.Text
	ChannelID      pgtype.Text
	Options        pgtype.JSONB
	StudentAnswers pgtype.JSONB
	StoppedAt      pgtype.Timestamptz
	EndedAt        pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func NewLiveRoomPollFromEntity(l *domain.LiveRoomPoll) (*LiveRoomPoll, error) {
	dto := &LiveRoomPoll{}
	database.AllNullEntity(dto)

	if err := multierr.Combine(
		dto.ChannelID.Set(l.ChannelID),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not map live room entity to dto: %w", err)
	}

	if len(l.StudentAnswers) > 0 {
		if err := dto.StudentAnswers.Set(database.JSONB(l.StudentAnswers)); err != nil {
			return nil, fmt.Errorf("could not map live room poll entity to dto, student answers: %w", err)
		}
	}

	if l.Options != nil {
		if err := dto.Options.Set(database.JSONB(l.Options)); err != nil {
			return nil, fmt.Errorf("could not map live room poll entity to dto, options: %w", err)
		}
	}

	if l.StoppedAt != nil {
		if err := dto.StoppedAt.Set(l.StoppedAt); err != nil {
			return nil, fmt.Errorf("could not map live room poll entity to dto, stopped at: %w", err)
		}
	}

	if l.EndedAt != nil {
		if err := dto.EndedAt.Set(l.EndedAt); err != nil {
			return nil, fmt.Errorf("could not map live room poll entity to dto, ended at: %w", err)
		}
	}

	if l.DeletedAt != nil {
		if err := dto.DeletedAt.Set(l.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not map live room poll entity to dto, deleted at: %w", err)
		}
	}

	return dto, nil
}

func (l *LiveRoomPoll) FieldMap() ([]string, []interface{}) {
	return []string{
			"live_room_poll_id",
			"channel_id",
			"options",
			"students_answers",
			"stopped_at",
			"ended_at",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.LiveRoomPollID,
			&l.ChannelID,
			&l.Options,
			&l.StudentAnswers,
			&l.StoppedAt,
			&l.EndedAt,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LiveRoomPoll) TableName() string {
	return "live_room_poll"
}

func (l *LiveRoomPoll) PreInsert() error {
	if l.LiveRoomPollID.Status != pgtype.Present {
		if err := l.LiveRoomPollID.Set(idutil.ULIDNow()); err != nil {
			return fmt.Errorf("failed to set ID in PreInsert: %w", err)
		}
	}

	return nil
}

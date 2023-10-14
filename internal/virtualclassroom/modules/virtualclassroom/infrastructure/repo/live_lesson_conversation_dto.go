package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveLessonConversation struct {
	LessonConversationID pgtype.Text
	ConversationID       pgtype.Text
	LessonID             pgtype.Text
	ParticipantList      pgtype.TextArray
	ConversationType     pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
}

func NewLiveLessonConversationDTO(l domain.LiveLessonConversation) (*LiveLessonConversation, error) {
	dto := &LiveLessonConversation{}
	database.AllNullEntity(dto)

	if err := multierr.Combine(
		dto.ConversationID.Set(l.ConversationID),
		dto.LessonID.Set(l.LessonID),
		dto.ParticipantList.Set(l.ParticipantList),
		dto.ConversationType.Set(string(l.ConversationType)),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from live lesson conversation domain to dto: %w", err)
	}

	return dto, nil
}

func (l *LiveLessonConversation) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_conversation_id",
			"conversation_id",
			"lesson_id",
			"participant_list",
			"conversation_type",
			"created_at",
			"updated_at",
		}, []interface{}{
			&l.LessonConversationID,
			&l.ConversationID,
			&l.LessonID,
			&l.ParticipantList,
			&l.ConversationType,
			&l.CreatedAt,
			&l.UpdatedAt,
		}
}

func (l *LiveLessonConversation) TableName() string {
	return "live_lesson_conversation"
}

func (l *LiveLessonConversation) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("failed to set values in PreInsert: %w", err)
	}

	if l.LessonConversationID.Status != pgtype.Present {
		if err := l.LessonConversationID.Set(idutil.ULIDNow()); err != nil {
			return fmt.Errorf("failed to set ID in PreInsert: %w", err)
		}
	}

	return nil
}

func (l *LiveLessonConversation) ToLiveLessonConversationDomain() domain.LiveLessonConversation {
	participants := make([]string, 0, len(l.ParticipantList.Elements))
	for _, participant := range l.ParticipantList.Elements {
		participants = append(participants, participant.String)
	}

	domain := domain.LiveLessonConversation{
		LessonConversationID: l.LessonConversationID.String,
		ConversationID:       l.ConversationID.String,
		LessonID:             l.LessonID.String,
		ParticipantList:      participants,
		ConversationType:     domain.LiveLessonConversationType(l.ConversationType.String),
		CreatedAt:            l.CreatedAt.Time,
		UpdatedAt:            l.UpdatedAt.Time,
	}

	return domain
}

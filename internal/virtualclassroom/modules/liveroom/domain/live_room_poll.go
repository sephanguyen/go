package domain

import (
	"time"

	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type LiveRoomPoll struct {
	LiveRoomPollID string
	ChannelID      string
	StudentAnswers StudentAnswersList
	StoppedAt      *time.Time
	EndedAt        *time.Time
	Options        *vc_domain.CurrentPollingOptions
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

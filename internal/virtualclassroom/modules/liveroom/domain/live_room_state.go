package domain

import (
	"time"

	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type LiveRoomState struct {
	LiveRoomStateID     string
	ChannelID           string
	CurrentMaterial     *vc_domain.CurrentMaterial
	SpotlightedUser     string
	WhiteboardZoomState *vc_domain.WhiteboardZoomState
	Recording           *vc_domain.CompositeRecordingState
	CurrentPolling      *vc_domain.CurrentPolling
	SessionTime         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

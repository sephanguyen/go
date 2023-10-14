package commands

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
)

type UpsertRecordingStatePayload struct {
	RecordingRef     constant.RecordingReference
	RecordingChannel string
	Recording        *domain.CompositeRecordingState
}

type NewRecordingVideoPayload struct {
	RecordingRef     constant.RecordingReference
	RecordingChannel string
	RecordedVideos   []*domain.RecordedVideo
}

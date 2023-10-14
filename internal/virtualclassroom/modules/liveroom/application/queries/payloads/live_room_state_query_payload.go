package payloads

import (
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type GetLiveRoomStateResponse struct {
	ChannelID     string
	LiveRoomState *domain.LiveRoomState
	Media         *media_domain.Media
	UserStates    *vc_domain.UserStates
}

type CreateAndGetChannelInfoResponse struct {
	ChannelID       string
	RoomID          string
	WhiteboardAppID string
	WhiteboardToken string
}

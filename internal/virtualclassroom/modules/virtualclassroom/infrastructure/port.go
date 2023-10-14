package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type VirtualClassroomPort interface {
	domain.VirtualClassroomPort
}

type LessonPort interface {
	domain.VirtualLessonPort
}

type StreamingProvider interface {
	domain.StreamingProviderPort
}

type UserModulePort interface {
	domain.UserModulePort
}

type WhiteboardPort interface {
	FetchRoomToken(ctx context.Context, roomUUID string) (string, error)
	CreateRoom(context.Context, *whiteboard.CreateRoomRequest) (*whiteboard.CreateRoomResponse, error)
}

type AgoraTokenPort interface {
	GenerateAgoraStreamToken(referenceID string, userID string, role domain.AgoraRole) (string, error)
	BuildRTMToken(referenceID string, userID string) (string, error)
	BuildRTMTokenByUserID(userID string) (string, error)
}

type MediaModulePort interface {
	RetrieveMediasByIDs(ctx context.Context, mediaIDs []string) (media_domain.Medias, error)
	CreateMedia(ctx context.Context, media *media_domain.Media) error
	DeleteMedias(ctx context.Context, mediaIDs []string) error
}

package queries

import (
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type LessonRoomStateQueryPayload struct {
	LessonID string
}

type GetLiveLessonStateResponse struct {
	LessonID        string
	Media           *media_domain.Media
	LessonRoomState *domain.LessonRoomState
	UserStates      *domain.UserStates
}

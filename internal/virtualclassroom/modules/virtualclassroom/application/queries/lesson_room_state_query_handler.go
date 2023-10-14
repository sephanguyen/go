package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type LessonRoomStateQuery struct {
	WrapperDBConnection *support.WrapperDBConnection
	VirtualLessonRepo   infrastructure.VirtualLessonRepo
	LessonRoomStateRepo infrastructure.LessonRoomStateRepo
	LessonMemberRepo    infrastructure.LessonMemberRepo
	MediaModulePort     infrastructure.MediaModulePort
	StudentsRepo        infrastructure.StudentsRepo
}

func (l *LessonRoomStateQuery) GetLessonRoomStateByLessonID(ctx context.Context, payload LessonRoomStateQueryPayload) (*domain.LessonRoomState, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	state, err := l.LessonRoomStateRepo.GetLessonRoomStateByLessonID(ctx, conn, payload.LessonID)

	if err == domain.ErrLessonRoomStateNotFound {
		t := &domain.LessonRoomState{
			Recording: &domain.CompositeRecordingState{},
		}
		return t, nil
	}
	return state, err
}

func (l *LessonRoomStateQuery) GetLessonRoomStateByLessonIDWithoutCheck(ctx context.Context, payload LessonRoomStateQueryPayload) (*domain.LessonRoomState, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	state, err := l.LessonRoomStateRepo.GetLessonRoomStateByLessonID(ctx, conn, payload.LessonID)
	return state, err
}

func (l *LessonRoomStateQuery) GetLiveLessonState(ctx context.Context, payload LessonRoomStateQueryPayload) (*GetLiveLessonStateResponse, error) {
	lessonID := payload.LessonID
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	lesson, err := l.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID)
	if err != nil {
		return nil, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", lessonID, err)
	}

	userID := interceptors.UserIDFromContext(ctx)
	if !lesson.TeacherIDs.HaveID(userID) && !lesson.LearnerIDs.HaveID(userID) {
		isUserAStudent, err := l.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
		if err != nil {
			return nil, fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
		}

		if isUserAStudent {
			return nil, fmt.Errorf("permission denied: user %s is a student who is not part of the lesson", userID)
		}
	}

	lessonRoomState, err := l.LessonRoomStateRepo.GetLessonRoomStateByLessonID(ctx, conn, lessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return nil, fmt.Errorf("error in LessonRoomStateRepo.GetLessonRoomStateByLessonID, lesson %s: %w", lessonID, err)
	}
	if err == domain.ErrLessonRoomStateNotFound {
		lessonRoomState = &domain.LessonRoomState{
			SpotlightedUser:     "",
			Recording:           &domain.CompositeRecordingState{},
			WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
		}
	}

	// get media of current material
	var media *media_domain.Media
	if lessonRoomState.CurrentMaterial != nil && lessonRoomState.CurrentMaterial.MediaID != "" {
		mediaID := lessonRoomState.CurrentMaterial.MediaID
		medias, err := l.MediaModulePort.RetrieveMediasByIDs(ctx, []string{mediaID})
		if err != nil {
			return nil, fmt.Errorf("error in MediaModulePort.RetrieveMediasByIDs, lesson %s, media %s: %w", lessonID, mediaID, err)
		}

		if len(medias) == 0 {
			return nil, fmt.Errorf("media %s is expected with current material of lesson %s but nothing found", lessonID, mediaID)
		}

		media = medias[0]
	}

	userState, err := l.GetLearnerStatesByLessonID(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &GetLiveLessonStateResponse{
		LessonID:        lessonID,
		Media:           media,
		LessonRoomState: lessonRoomState,
		UserStates:      userState,
	}, nil
}

func (l *LessonRoomStateQuery) GetLearnerStatesByLessonID(ctx context.Context, payload LessonRoomStateQueryPayload) (*domain.UserStates, error) {
	var userStates *domain.UserStates
	lessonID := payload.LessonID
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	learnerStates, err := l.LessonMemberRepo.GetLessonMemberStatesByLessonID(ctx, conn, lessonID)
	if err != nil {
		return userStates, fmt.Errorf("error in LessonMemberRepo.GetLessonMemberStatesByLessonID, lesson %s: %w", lessonID, err)
	}

	for _, learnerState := range learnerStates {
		if learnerState.LessonID != lessonID {
			return userStates, fmt.Errorf(`learner %s state not belong to lesson %s, got state for lesson %s`, learnerState.UserID, lessonID, learnerState.LessonID)
		}
	}

	userStates = domain.NewUserState(learnerStates)
	return userStates, nil
}

package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

// command processors

type CommandProcessor interface {
	Execute(context.Context, StateModifyCommand) error
}

var (
	_ CommandProcessor = new(CommandDispatcher)
	_ CommandProcessor = new(CommandPermissionChecker)
)

type CommandDispatcher struct {
	cp CommandProcessor

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
		GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
		StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
	LessonGroupRepo interface {
		Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error)
	}
	LessonPollingRepo interface {
		Create(ctx context.Context, db database.Ext, polling *entities.LessonPolling) (*entities.LessonPolling, error)
	}
	LessonRoomStateRepo interface {
		Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error
		UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
		UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error
		UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error
	}
}

func (dp *CommandDispatcher) Execute(ctx context.Context, command StateModifyCommand) error {
	if dp.cp != nil {
		if err := dp.cp.Execute(ctx, command); err != nil {
			return err
		}
	}

	var handler CommandHandler
	switch v := command.(type) {
	case *ShareMaterialCommand:
		handler = &ShareMaterialCommandHandler{
			command:             v,
			DB:                  dp.DB,
			LessonGroupRepo:     dp.LessonGroupRepo,
			LessonRepo:          dp.LessonRepo,
			LessonMemberRepo:    dp.LessonMemberRepo,
			LessonRoomStateRepo: dp.LessonRoomStateRepo,
		}
	case *StopSharingMaterialCommand:
		handler = &StopSharingMaterialCommandHandler{
			command:             v,
			DB:                  dp.DB,
			LessonRepo:          dp.LessonRepo,
			LessonMemberRepo:    dp.LessonMemberRepo,
			LessonRoomStateRepo: dp.LessonRoomStateRepo,
		}
	case *FoldHandAllCommand:
		handler = &FoldHandAllCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *UpdateHandsUpCommand:
		handler = &UpdateHandsUpCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *ResetAllStatesCommand:
		handler = &ResetAllStatesCommandHandler{
			command:             v,
			DB:                  dp.DB,
			LessonRepo:          dp.LessonRepo,
			LessonMemberRepo:    dp.LessonMemberRepo,
			LessonRoomStateRepo: dp.LessonRoomStateRepo,
		}
	case *UpdateAnnotationCommand:
		handler = &UpdateAnnotationCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonRepo:       dp.LessonRepo,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *DisableAllAnnotationCommand:
		handler = &DisableAllAnnotationCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *StartPollingCommand:
		handler = &StartPollingCommandHandler{
			command:    v,
			DB:         dp.DB,
			LessonRepo: dp.LessonRepo,
		}
	case *StopPollingCommand:
		handler = &StopPollingCommandHandler{
			command:    v,
			DB:         dp.DB,
			LessonRepo: dp.LessonRepo,
		}
	case *EndPollingCommand:
		handler = &EndPollingCommandHandler{
			command:           v,
			DB:                dp.DB,
			LessonRepo:        dp.LessonRepo,
			LessonMemberRepo:  dp.LessonMemberRepo,
			LessonPollingRepo: dp.LessonPollingRepo,
		}
	case *SubmitPollingAnswerCommand:
		handler = &SubmitPollingAnswerCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonRepo:       dp.LessonRepo,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *ResetPollingCommand:
		handler = &ResetPollingCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonRepo:       dp.LessonRepo,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *RequestRecordingCommand:
		handler = &RequestRecordingHandler{
			command:    v,
			DB:         dp.DB,
			LessonRepo: dp.LessonRepo,
		}
	case *StopRecordingCommand:
		handler = &StopRecordingHandler{
			command:    v,
			DB:         dp.DB,
			LessonRepo: dp.LessonRepo,
		}
	case *SpotlightCommand:
		handler = &SpotlightCommandHandler{
			command:             v,
			DB:                  dp.DB,
			LessonRoomStateRepo: dp.LessonRoomStateRepo,
		}
	case *WhiteboardZoomStateCommand:
		handler = &WhiteboardZoomStateCommandHandler{
			command:             v,
			DB:                  dp.DB,
			LessonRoomStateRepo: dp.LessonRoomStateRepo,
		}
	case *UpdateChatCommand:
		handler = &UpdateChatCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonRepo:       dp.LessonRepo,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	case *ResetAllChatCommand:
		handler = &ResetAllChatCommandHandler{
			command:          v,
			DB:               dp.DB,
			LessonMemberRepo: dp.LessonMemberRepo,
		}
	default:
		return fmt.Errorf("unimplement handler of command type %T", command)
	}

	if err := handler.Execute(ctx); err != nil {
		return err
	}

	return nil
}

type CommandPermissionChecker struct {
	cp     CommandProcessor
	lesson *entities.Lesson

	DB       database.Ext
	UserRepo interface {
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
	}
}

func (dp *CommandPermissionChecker) Execute(ctx context.Context, command StateModifyCommand) error {
	if dp.cp != nil {
		if err := dp.cp.Execute(ctx, command); err != nil {
			return err
		}
	}

	isTeacher := func(commander pgtype.Text) error {
		if !dp.lesson.TeacherIDs.HaveID(commander) {
			userGroup, err := dp.UserRepo.UserGroup(ctx, dp.DB, commander)
			if err != nil {
				return fmt.Errorf("UserRepo.UserGroup: %w", err)
			}

			if userGroup == entities.UserGroupStudent {
				return fmt.Errorf("permission denied: user %s can not execute this command", commander.String)
			}
		}

		return nil
	}

	commander := command.getCommander()
	switch v := command.(type) {
	case *UpdateHandsUpCommand:
		if dp.lesson.LearnerIDs.HaveID(database.Text(commander)) {
			// learner can only update self-state
			if commander != v.UserID {
				return fmt.Errorf("permission denied: learner %s can not change other member's status", commander)
			}
		} else if err := isTeacher(database.Text(commander)); err != nil {
			return err
		}
	case *SubmitPollingAnswerCommand:
		if dp.lesson.LearnerIDs.HaveID(database.Text(commander)) {
			// learner can only submit self-state
			if commander != v.UserID {
				return fmt.Errorf("permission denied: other learner %s can not submit polling answer", commander)
			}
		} else {
			return fmt.Errorf("permission denied: user %s can not submit polling answer", commander)
		}
	default:
		if err := isTeacher(database.Text(commander)); err != nil {
			return err
		}
	}

	return nil
}

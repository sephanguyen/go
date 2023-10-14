package commands

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type CommandHandler interface {
	Execute() error
}

func GetCommandHandler(m *ModifyLiveRoomCommandDispatcher, fGetLessonmgmtDB func() database.Ext, command ModifyStateCommand) (CommandHandler, error) {
	var handler CommandHandler
	db := fGetLessonmgmtDB()

	switch c := command.(type) {
	case *UpdateAnnotationCommand:
		handler = &UpdateAnnotationCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *EnableAllAnnotationCommand:
		handler = &EnableAllAnnotationCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *DisableAllAnnotationCommand:
		handler = &DisableAllAnnotationCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *UpdateChatCommand:
		handler = &UpdateChatCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *ResetAllChatCommand:
		handler = &ResetAllChatCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *StartPollingCommand:
		handler = &StartPollingCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *StopPollingCommand:
		handler = &StopPollingCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *EndPollingCommand:
		handler = &EndPollingCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomStateRepo:       m.LiveRoomStateRepo,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
			LiveRoomPoll:            m.LiveRoomPoll,
		}
	case *SubmitPollingAnswerCommand:
		handler = &SubmitPollingAnswerCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomStateRepo:       m.LiveRoomStateRepo,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *SharePollingCommand:
		handler = &SharePollingCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *ResetPollingCommand:
		handler = &ResetPollingCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomStateRepo:       m.LiveRoomStateRepo,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *UpdateHandsUpCommand:
		handler = &UpdateHandsUpCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *FoldHandAllCommand:
		handler = &FoldHandAllCommandHandler{
			command:                 c,
			lessonmgmtDB:            db,
			ctx:                     m.Ctx,
			LiveRoomMemberStateRepo: m.LiveRoomMemberStateRepo,
		}
	case *SpotlightCommand:
		handler = &SpotlightCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *WhiteboardZoomStateCommand:
		handler = &WhiteboardZoomStateCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *ShareMaterialCommand:
		handler = &ShareMaterialCommandHandler{
			command:       c,
			lessonmgmtDB:  db,
			ctx:           m.Ctx,
			dispatcher:    m,
			LiveRoomState: m.LiveRoomStateRepo,
		}
	case *StopSharingMaterialCommand:
		handler = &StopSharingMaterialCommandHandler{
			command:      c,
			lessonmgmtDB: db,
			ctx:          m.Ctx,
			dispatcher:   m,
		}
	case *UpsertSessionTimeCommand:
		handler = &UpsertSessionTimeCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *ClearRecordingCommand:
		handler = &ClearRecordingCommandHandler{
			command:           c,
			lessonmgmtDB:      db,
			ctx:               m.Ctx,
			LiveRoomStateRepo: m.LiveRoomStateRepo,
		}
	case *ResetAllStatesCommand:
		handler = &ResetAllStatesCommandHandler{
			command:      c,
			lessonmgmtDB: db,
			ctx:          m.Ctx,
			dispatcher:   m,
		}
	default:
		return nil, fmt.Errorf("unimplemented handler of command type %T", command)
	}
	return handler, nil
}

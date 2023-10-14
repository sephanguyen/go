package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type VirtualClassroomDispatcherConfig struct {
	Ctx                      context.Context
	LessonGroupRepo          infrastructure.LessonGroupRepo
	VirtualLessonRepo        infrastructure.VirtualLessonRepo
	LessonMemberRepo         infrastructure.LessonMemberRepo
	VirtualLessonPollingRepo infrastructure.VirtualLessonPollingRepo
	LessonRoomStateRepo      infrastructure.LessonRoomStateRepo
	DB                       database.Ext
	PermissionChecker        CommandChecker
}

type VirtualClassroomDispatcher struct {
	*VirtualClassroomDispatcherConfig
}

type Dispatcher interface {
	Dispatch(command StateModifyCommand) error
	DispatchWithTransaction(trx database.Ext, command StateModifyCommand) error
}

func NewDispatcher(cf *VirtualClassroomDispatcherConfig) *VirtualClassroomDispatcher {
	return &VirtualClassroomDispatcher{
		&VirtualClassroomDispatcherConfig{
			Ctx:                      cf.Ctx,
			LessonGroupRepo:          cf.LessonGroupRepo,
			VirtualLessonRepo:        cf.VirtualLessonRepo,
			LessonMemberRepo:         cf.LessonMemberRepo,
			DB:                       cf.DB,
			VirtualLessonPollingRepo: cf.VirtualLessonPollingRepo,
			PermissionChecker:        cf.PermissionChecker,
			LessonRoomStateRepo:      cf.LessonRoomStateRepo,
		},
	}
}

func getHandler(dispatcher *VirtualClassroomDispatcher, fGetDB func() database.Ext, command StateModifyCommand) (CommandHandler, error) {
	var handler CommandHandler
	db := fGetDB()
	switch v := command.(type) {
	case *ShareMaterialCommand:
		handler = &ShareMaterialCommandHandler{
			db:                  db,
			ctx:                 dispatcher.Ctx,
			lessonGroupRepo:     dispatcher.LessonGroupRepo,
			virtualLessonRepo:   dispatcher.VirtualLessonRepo,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
			command:             v,
			dispatcher:          dispatcher,
		}
	case *StopSharingMaterialCommand:
		handler = &StopSharingMaterialCommandHandler{
			command:    v,
			ctx:        dispatcher.Ctx,
			db:         db,
			dispatcher: dispatcher,
		}
	case *FoldHandAllCommand:
		handler = &FoldHandAllCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *UpdateHandsUpCommand:
		handler = &UpdateHandsUpCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *UpdateAnnotationCommand:
		handler = &UpdateAnnotationCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *DisableAllAnnotationCommand:
		handler = &DisableAllAnnotationCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *StartPollingCommand:
		handler = &StartPollingCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *StopPollingCommand:
		handler = &StopPollingCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *EndPollingCommand:
		handler = &EndPollingCommandHandler{
			command:                  v,
			ctx:                      dispatcher.Ctx,
			db:                       db,
			lessonMemberRepo:         dispatcher.LessonMemberRepo,
			virtualLessonPollingRepo: dispatcher.VirtualLessonPollingRepo,
			lessonRoomStateRepo:      dispatcher.LessonRoomStateRepo,
		}
	case *SubmitPollingAnswerCommand:
		handler = &SubmitPollingAnswerCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonMemberRepo:    dispatcher.LessonMemberRepo,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *SharePollingCommand:
		handler = &SharePollingCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *ResetPollingCommand:
		handler = &ResetPollingCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonMemberRepo:    dispatcher.LessonMemberRepo,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *ResetAllStatesCommand:
		handler = &ResetAllStatesCommandHandler{
			command:    v,
			ctx:        dispatcher.Ctx,
			db:         db,
			dispatcher: dispatcher,
		}
	case *WhiteboardZoomStateCommand:
		handler = &WhiteboardZoomStateCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *SpotlightCommand:
		handler = &SpotlightCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *UpdateChatCommand:
		handler = &UpdateChatCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *UpsertSessionTimeCommand:
		handler = &UpsertSessionTimeCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	case *ResetAllChatCommand:
		handler = &ResetAllChatCommandHandler{
			command:          v,
			ctx:              dispatcher.Ctx,
			db:               db,
			lessonMemberRepo: dispatcher.LessonMemberRepo,
		}
	case *ClearRecordingCommand:
		handler = &ClearRecordingCommandHandler{
			command:             v,
			ctx:                 dispatcher.Ctx,
			db:                  db,
			lessonRoomStateRepo: dispatcher.LessonRoomStateRepo,
		}
	default:
		return nil, fmt.Errorf("unimplemented handler of command type %T", command)
	}
	return handler, nil
}

func (v *VirtualClassroomDispatcher) pDispatch(fGetDB func() database.Ext, command StateModifyCommand) error {
	handler, err := getHandler(v, fGetDB, command)
	if err != nil {
		return err
	}
	if err := handler.Execute(); err != nil {
		return err
	}
	return nil
}

func (v *VirtualClassroomDispatcher) Dispatch(command StateModifyCommand) error {
	fGetDB := func() database.Ext {
		return v.DB
	}
	return v.pDispatch(fGetDB, command)
}

func (v *VirtualClassroomDispatcher) CheckPermissionAndDispatch(command StateModifyCommand) error {
	if errPermission := v.PermissionChecker.Check(command); errPermission != nil {
		return errPermission
	}
	return v.Dispatch(command)
}

func (v *VirtualClassroomDispatcher) DispatchWithTransaction(trx database.Ext, command StateModifyCommand) error {
	fGetDB := func() database.Ext {
		return trx
	}
	return v.pDispatch(fGetDB, command)
}

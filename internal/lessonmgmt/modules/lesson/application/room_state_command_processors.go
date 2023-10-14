package application

import (
	"context"
	"fmt"

	entitiesBob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type StateModifyCommand interface {
	GetCommander() string
	GetLessonID() string
}

type CommandHandler interface {
	Execute(ctx context.Context) error
}

// command processors

type RoomStateCommandProcessor interface {
	Execute(context.Context, StateModifyCommand) error
}

var (
	_ RoomStateCommandProcessor = new(RoomStateCommandDispatcher)
	_ RoomStateCommandProcessor = new(RoomStateCommandPermissionChecker)
)

type RoomStateCommandDispatcher struct {
	WrapperConnection *support.WrapperDBConnection
	LessonRepo        infrastructure.LessonRepo
	MediaModulePort   infrastructure.MediaModulePort
	RoomStateRepo     infrastructure.LessonRoomState
}

func (dp *RoomStateCommandDispatcher) Execute(ctx context.Context, command StateModifyCommand) error {
	var handler CommandHandler
	switch v := command.(type) {
	case *ModifyCurrentMaterialCommand:
		handler = &ModifyCurrentMaterialCommandHandler{
			command:           v,
			WrapperConnection: dp.WrapperConnection,
			LessonRepo:        dp.LessonRepo,
			MediaModulePort:   dp.MediaModulePort,
			RoomStateRepo:     dp.RoomStateRepo,
		}
	case *ShareMaterialCommand:
		handler = &ShareMaterialCommandHandler{
			command:           v,
			WrapperConnection: dp.WrapperConnection,
			LessonRepo:        dp.LessonRepo,
			MediaModulePort:   dp.MediaModulePort,
			RoomStateRepo:     dp.RoomStateRepo,
		}
	case *StopSharingMaterialCommand:
		handler = &StopSharingMaterialCommandHandler{
			command:           v,
			WrapperConnection: dp.WrapperConnection,
			LessonRepo:        dp.LessonRepo,
			RoomStateRepo:     dp.RoomStateRepo,
		}
	default:
		return fmt.Errorf("unimplement handler of command type %T", command)
	}

	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("got error when execure handler: %w", err)
	}

	return nil
}

type RoomStateCommandPermissionChecker struct {
	WrapperConnection *support.WrapperDBConnection
	UserModule        infrastructure.UserModulePort
}

func (dp *RoomStateCommandPermissionChecker) Execute(ctx context.Context, command StateModifyCommand) error {
	isTeacher := func(commander string) error {
		userGroup, err := dp.UserModule.GetUserGroup(ctx, commander)
		if err != nil {
			return fmt.Errorf("UserRepo.GetUserGroup: %w", err)
		}

		if userGroup == entitiesBob.UserGroupStudent {
			return fmt.Errorf("permission denied: user %s can not execute this command", commander)
		}

		return nil
	}

	commander := command.GetCommander()
	if err := isTeacher(commander); err != nil {
		return err
	}

	return nil
}

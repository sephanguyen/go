package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"

	"github.com/jackc/pgx/v4"
)

var (
	_ StateModifyCommand = new(ModifyCurrentMaterialCommand)
	_ StateModifyCommand = new(ShareMaterialCommand)
	_ StateModifyCommand = new(StopSharingMaterialCommand)
)

type ModifyCurrentMaterialCommand struct {
	CommanderID string
	LessonID    string
	MediaID     *string
	VideoState  *domain.VideoState
}

func (s *ModifyCurrentMaterialCommand) GetCommander() string {
	return s.CommanderID
}

func (s *ModifyCurrentMaterialCommand) GetLessonID() string {
	return s.LessonID
}

type ShareMaterialCommand struct {
	CommanderID string
	LessonID    string
	MediaID     string
	VideoState  *domain.VideoState
}

func (s *ShareMaterialCommand) GetCommander() string {
	return s.CommanderID
}

func (s *ShareMaterialCommand) GetLessonID() string {
	return s.LessonID
}

type StopSharingMaterialCommand struct {
	CommanderID string
	LessonID    string
}

func (s *StopSharingMaterialCommand) GetCommander() string {
	return s.CommanderID
}

func (s *StopSharingMaterialCommand) GetLessonID() string {
	return s.LessonID
}

// List command handlers for above commands
var _ CommandHandler = new(ModifyCurrentMaterialCommandHandler)
var _ CommandHandler = new(ShareMaterialCommandHandler)
var _ CommandHandler = new(StopSharingMaterialCommandHandler)

type ModifyCurrentMaterialCommandHandler struct {
	command           *ModifyCurrentMaterialCommand
	WrapperConnection *support.WrapperDBConnection

	// ports
	LessonRepo      infrastructure.LessonRepo
	MediaModulePort infrastructure.MediaModulePort
	RoomStateRepo   infrastructure.LessonRoomState
}

func (s *ModifyCurrentMaterialCommandHandler) execute(ctx context.Context, db database.Ext) error {
	builder := domain.NewCurrentMaterial().
		WithLessonID(s.command.LessonID).
		WithLessonRepo(s.LessonRepo).
		WithMediaModulePort(s.MediaModulePort)
	if s.command.MediaID != nil {
		builder.WithMediaID(*s.command.MediaID)
	}
	if s.command.VideoState != nil {
		builder.
			WithVideoCurrentTime(s.command.VideoState.CurrentTime).
			WithVideoPlayerState(s.command.VideoState.PlayerState)
	}

	currentMatl, err := builder.Build(ctx, db)
	if err != nil {
		return err
	}
	currentMatl.PreInsert()
	_, err = s.RoomStateRepo.UpsertCurrentMaterial(ctx, db, currentMatl)
	if err != nil {
		return fmt.Errorf("RoomStateRepo.UpsertCurrentMaterial: %w", err)
	}

	return nil
}

func (s *ModifyCurrentMaterialCommandHandler) Execute(ctx context.Context) error {
	conn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil
	}
	return database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		return s.execute(ctx, tx)
	})
}

type ShareMaterialCommandHandler struct {
	command           *ShareMaterialCommand
	WrapperConnection *support.WrapperDBConnection

	// ports
	LessonRepo      infrastructure.LessonRepo
	MediaModulePort infrastructure.MediaModulePort
	RoomStateRepo   infrastructure.LessonRoomState
}

func (s *ShareMaterialCommandHandler) Execute(ctx context.Context) error {
	command := ModifyCurrentMaterialCommand{
		CommanderID: s.command.CommanderID,
		LessonID:    s.command.LessonID,
		MediaID:     &s.command.MediaID,
		VideoState:  s.command.VideoState,
	}
	commandDp := RoomStateCommandDispatcher{
		WrapperConnection: s.WrapperConnection,
		LessonRepo:        s.LessonRepo,
		MediaModulePort:   s.MediaModulePort,
		RoomStateRepo:     s.RoomStateRepo,
	}
	if err := commandDp.Execute(ctx, &command); err != nil {
		return err
	}

	// TODO: Reset annotation states when share video

	return nil
}

type StopSharingMaterialCommandHandler struct {
	command           *StopSharingMaterialCommand
	WrapperConnection *support.WrapperDBConnection

	// ports
	LessonRepo    infrastructure.LessonRepo
	RoomStateRepo infrastructure.LessonRoomState
}

func (s *StopSharingMaterialCommandHandler) Execute(ctx context.Context) error {
	command := ModifyCurrentMaterialCommand{
		CommanderID: s.command.CommanderID,
		LessonID:    s.command.LessonID,
	}
	commandDp := RoomStateCommandDispatcher{
		WrapperConnection: s.WrapperConnection,
		LessonRepo:        s.LessonRepo,
		RoomStateRepo:     s.RoomStateRepo,
	}
	if err := commandDp.Execute(ctx, &command); err != nil {
		return err
	}

	// TODO: Reset annotation states when stop sharing

	return nil
}

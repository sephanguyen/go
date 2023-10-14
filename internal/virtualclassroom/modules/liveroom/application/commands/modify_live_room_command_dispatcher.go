package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
)

type ModifyLiveRoomCommandDispatcherConfig struct {
	Ctx               context.Context
	LessonmgmtDB      database.Ext
	PermissionChecker CommandChecker

	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
	LiveRoomPoll            infrastructure.LiveRoomPoll
}

type ModifyLiveRoomCommandDispatcher struct {
	*ModifyLiveRoomCommandDispatcherConfig
}

type Dispatcher interface {
	Dispatch(command ModifyStateCommand) error
	DispatchWithTransaction(trx database.Ext, command ModifyStateCommand) error
}

func NewDispatcher(config *ModifyLiveRoomCommandDispatcherConfig) *ModifyLiveRoomCommandDispatcher {
	return &ModifyLiveRoomCommandDispatcher{
		&ModifyLiveRoomCommandDispatcherConfig{
			Ctx:                     config.Ctx,
			LessonmgmtDB:            config.LessonmgmtDB,
			PermissionChecker:       config.PermissionChecker,
			LiveRoomStateRepo:       config.LiveRoomStateRepo,
			LiveRoomMemberStateRepo: config.LiveRoomMemberStateRepo,
			LiveRoomPoll:            config.LiveRoomPoll,
		},
	}
}

func (m *ModifyLiveRoomCommandDispatcher) pDispatch(fGetLessonmgmtDB func() database.Ext, command ModifyStateCommand) error {
	handler, err := GetCommandHandler(m, fGetLessonmgmtDB, command)
	if err != nil {
		return err
	}
	if err := handler.Execute(); err != nil {
		return err
	}
	return nil
}

func (m *ModifyLiveRoomCommandDispatcher) Dispatch(command ModifyStateCommand) error {
	fGetLessonmgmtDB := func() database.Ext {
		return m.LessonmgmtDB
	}
	return m.pDispatch(fGetLessonmgmtDB, command)
}

func (m *ModifyLiveRoomCommandDispatcher) CheckPermissionAndDispatch(command ModifyStateCommand) error {
	if errPermission := m.PermissionChecker.Check(command); errPermission != nil {
		return errPermission
	}
	return m.Dispatch(command)
}

func (m *ModifyLiveRoomCommandDispatcher) DispatchWithTransaction(tx database.Ext, command ModifyStateCommand) error {
	fGetLessonmgmtDB := func() database.Ext {
		return tx
	}
	return m.pDispatch(fGetLessonmgmtDB, command)
}

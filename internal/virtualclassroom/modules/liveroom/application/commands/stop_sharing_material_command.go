package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

type StopSharingMaterialCommand struct {
	*ModifyLiveRoomCommand
}

type StopSharingMaterialCommandHandler struct {
	command      *StopSharingMaterialCommand
	ctx          context.Context
	lessonmgmtDB database.Ext
	dispatcher   Dispatcher
}

func (s *StopSharingMaterialCommandHandler) pExecute(db database.Ext) error {
	err := s.dispatcher.DispatchWithTransaction(db, &ShareMaterialCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: s.command.CommanderID,
			ChannelID:   s.command.ChannelID},
	})

	return err
}

func (s *StopSharingMaterialCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}

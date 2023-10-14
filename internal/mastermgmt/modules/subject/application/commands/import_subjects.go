package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ImportSubjectsCommandHandler struct {
	DB          database.Ext
	SubjectRepo infrastructure.SubjectRepo
}
type ImportSubjectsPayload struct {
	Subjects []*domain.Subject
}

func (i *ImportSubjectsCommandHandler) ImportSubjects(ctx context.Context, payload ImportSubjectsPayload) (err error) {
	err = database.ExecInTx(ctx, i.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = i.SubjectRepo.Import(ctx, i.DB, payload.Subjects)
		return err
	})

	if err != nil {
		return fmt.Errorf("SubjectRepo.Import: %w", err)
	}
	return nil
}

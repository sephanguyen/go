package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ImportGradesCommandHandler struct {
	DB        database.Ext
	GradeRepo infrastructure.GradeRepo
}

func (i *ImportGradesCommandHandler) ImportGrades(ctx context.Context, payload ImportGradesPayload) (err error) {
	partnerIDs := sliceutils.Map(payload.Grades, func(g *domain.Grade) string {
		return g.PartnerInternalID
	})

	egs, err := i.GradeRepo.GetByPartnerInternalIDs(ctx, i.DB, partnerIDs)
	if err != nil {
		return err
	}

	egMap := make(map[string]*domain.Grade, len(egs))
	for _, v := range egs {
		egMap[v.PartnerInternalID] = v
	}

	// if exists then get the existing id, if not, generate an id
	for _, g := range payload.Grades {
		eg, ok := egMap[g.PartnerInternalID]
		if ok {
			g.ID = eg.ID
		} else {
			g.ID = idutil.ULIDNow()
		}
	}

	err = database.ExecInTx(ctx, i.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = i.GradeRepo.Import(ctx, i.DB, payload.Grades)
		return err
	})

	if err != nil {
		return fmt.Errorf("GradeRepo.Import: %w", err)
	}
	return nil
}

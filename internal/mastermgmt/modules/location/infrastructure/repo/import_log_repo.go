package repo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
)

type ImportLogRepo struct{}

func (rcv *ImportLogRepo) Create(ctx context.Context, db database.QueryExecer, importLog *domain.ImportLog) error {
	ctx, span := interceptors.StartSpan(ctx, "ImportLogRepo.Create")
	defer span.End()
	importLogDto, err := ToImportLog(importLog)
	if err != nil {
		return fmt.Errorf("cannot generate importLog dto: %w", err)
	}
	_, err = database.Insert(ctx, importLogDto, db.Exec)
	if err != nil {
		return fmt.Errorf("cannot insert: %w", err)
	}
	return nil
}

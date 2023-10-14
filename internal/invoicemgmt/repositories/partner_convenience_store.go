package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type PartnerConvenienceStoreRepo struct {
}

func (r *PartnerConvenienceStoreRepo) FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerConvenienceStore, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerConvenienceStoreRepo.FindOne")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	e := &entities.PartnerConvenienceStore{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE is_archived = false AND deleted_at IS null AND resource_path = $1 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, resourcePath).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

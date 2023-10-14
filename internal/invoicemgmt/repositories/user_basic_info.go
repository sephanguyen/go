package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type UserBasicInfoRepo struct{}

func (r *UserBasicInfoRepo) FindByID(ctx context.Context, db database.QueryExecer, userID string) (*entities.UserBasicInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserBasicInfoRepo.FindByID")
	defer span.End()

	e := &entities.UserBasicInfo{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &userID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}

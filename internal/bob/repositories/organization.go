package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type OrganizationRepo struct{}

func (r *OrganizationRepo) GetOrganizations(ctx context.Context, db database.Ext) ([]string, error) {
	var ret pgtype.TextArray
	err := db.QueryRow(ctx, "select array_agg(organization_id) from organizations").Scan(&ret)
	if err != nil {
		return nil, err
	}
	return database.FromTextArray(ret), nil
}

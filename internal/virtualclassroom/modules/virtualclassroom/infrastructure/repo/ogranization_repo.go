package repo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type OrganizationRepo struct{}

func (o *OrganizationRepo) GetIDs(ctx context.Context, db database.QueryExecer) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.GetIDs")
	defer span.End()

	orgQuery := "select organization_id from organizations order by created_at"
	organizations, err := db.Query(ctx, orgQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization:%w", err)
	}
	defer organizations.Close()
	ids := []string{}
	for organizations.Next() {
		var id string
		err := organizations.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization:%w", err)
		}
		ids = append(ids, id)
	}
	if err := organizations.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
)

type LocationRepository struct{}

func (l *LocationRepository) GetLocationByID(ctx context.Context, db database.Ext, id []string) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepository.GetLocationByID")
	defer span.End()

	fields := database.GetFieldNames(&domain.Location{})
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE location_id = ANY($1)
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	row, err := db.Query(ctx, query, &id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	res := []*domain.Location{}
	for row.Next() {
		location := &domain.Location{}
		_, value := location.FieldMap()
		if err = row.Scan(value...); err != nil {
			return nil, err
		}
		res = append(res, location)
	}
	if err = row.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

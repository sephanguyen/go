package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LocationRepo struct{}

func (l *LocationRepo) FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.FindByID")
	defer span.End()

	location := &entities.Location{}
	fields, values := location.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE location_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return location, nil
}

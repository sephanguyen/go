package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type LocationRepo struct{}

func (l *LocationRepo) GetLocationByID(ctx context.Context, db database.QueryExecer, id string) (*dto.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByID")
	defer span.End()

	location := &Location{}
	fields, values := location.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE location_id = $1 AND deleted_at IS NULL `,
		strings.Join(fields, ","),
		location.TableName(),
	)
	if err := db.QueryRow(ctx, query, id).Scan(values...); err != nil {
		return nil, err
	}

	return location.ConvertToDTO(), nil
}

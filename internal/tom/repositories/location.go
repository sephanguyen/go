package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LocationRepo struct{}

func (l *LocationRepo) FindAccessPaths(ctx context.Context, db database.Ext, locationIDs []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.FindAccessPaths")
	defer span.End()

	query := `
		SELECT array_agg(access_path) FROM locations
		WHERE location_id = ANY($1)
			AND deleted_at IS NULL`

	ids := pgtype.TextArray{}

	err := db.QueryRow(ctx, query, locationIDs).Scan(&ids)
	if err != nil {
		return []string{}, fmt.Errorf("db.QueryRow: %w", err)
	}
	if ids.Status == pgtype.Null {
		return []string{}, nil
	}
	return database.FromTextArray(ids), nil
}

func (l *LocationRepo) FindRootIDs(ctx context.Context, db database.Ext) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.FindRootIDs")
	defer span.End()

	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT l.location_id FROM locations l JOIN location_types lt 
		    ON l.location_type = lt.location_type_id 
			WHERE lt."name" LIKE '%org%' AND lt.resource_path = $1 AND l.resource_path = $1
			AND l.deleted_at IS NULL AND lt.deleted_at IS NULL`

	rows, err := db.Query(ctx, query, resourcePath)
	if err != nil {
		return []string{}, fmt.Errorf("db.QueryRow: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id pgtype.Text
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("db.rows.Scan: %w", err)
		}
		ids = append(ids, id.String)
	}

	return ids, nil
}

func (l *LocationRepo) FindLowestAccessPathByLocationIDs(ctx context.Context, db database.Ext, locationIDs []string) ([]string, map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.FindLowestAccessPathByLocationIDs")
	defer span.End()

	query := `
		SELECT l1.location_id, l1.access_path
		FROM locations l
			JOIN locations l1 ON l1.access_path ~ l.access_path AND l.resource_path = l1.resource_path
		WHERE l.location_id = ANY($1) 
			AND l.deleted_at IS NULL
			AND l1.deleted_at IS NULL;
	`
	rows, err := db.Query(ctx, query, database.TextArray(locationIDs))
	if err != nil {
		return nil, nil, err
	}
	lowestLocations := make([]string, 0)
	locationMapAccessPath := make(map[string]string, 0)
	defer rows.Close()
	for rows.Next() {
		var accessPath pgtype.Text
		var locationID pgtype.Text
		if err = rows.Scan(&locationID, &accessPath); err != nil {
			return nil, nil, err
		}
		lowestLocations = append(lowestLocations, locationID.String)
		locationMapAccessPath[locationID.String] = accessPath.String
	}

	return lowestLocations, locationMapAccessPath, nil
}

package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
)

type LocationRepo struct{}

type GetGrantedLowestLevelLocationsParams struct {
	Name            string
	Limit           int32
	UserID          string
	PermissionNames []string
	LocationIDs     []string
}

func (r *LocationRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, locationID string) (entities.Location, error) {
	location := &entities.Location{}
	locationFieldNames, locationFieldValues := location.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			location_id = $1
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(locationFieldNames, ","),
		location.TableName(),
	)
	row := db.QueryRow(ctx, stmt, locationID)
	err := row.Scan(locationFieldValues...)
	if err != nil {
		return entities.Location{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *location, nil
}

func (r *LocationRepo) GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Location, error) {
	var locations []entities.Location
	locationFieldNames, _ := (&entities.Location{}).FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			location_id = ANY($1)
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(locationFieldNames, ","),
		(&entities.Location{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		location := new(entities.Location)
		_, fieldValues := location.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		locations = append(locations, *location)
	}
	return locations, nil
}

func (r *LocationRepo) GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Location, error) {
	location := &entities.Location{}
	locationFieldNames, locationFieldValues := location.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			location_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(locationFieldNames, ","),
		location.TableName(),
	)
	row := db.QueryRow(ctx, stmt, entitiesID)
	err := row.Scan(locationFieldValues...)
	if err != nil {
		return entities.Location{}, err
	}
	return *location, nil
}

func (r *LocationRepo) GetLowestGrantedLocationIDsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, params GetGrantedLowestLevelLocationsParams) (locationIDs []string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLowestGrantedLocationIDsByUserIDAndPermissions")
	defer span.End()

	query := `
		SELECT l.location_id 
		FROM locations l 
		JOIN location_types lt ON l.location_type  = lt.location_type_id 
		WHERE lt.level = (SELECT MAX(level) FROM location_types lt2 
							WHERE lt2.deleted_at IS NULL AND lt2.is_archived = FALSE)
			AND l.location_id IN (
			    select t.location_id from 
				(
					select gp.location_id, array_agg(gp.permission_name) as permission_arr 
					from granted_permissions gp 
					where gp.user_id = $1::TEXT
					group by gp.location_id
				) as t
				where $2::TEXT[] <@ permission_arr
			)
			AND lower(l.name) LIKE lower(CONCAT('%%',$3::text,'%%')) AND l.is_archived = false  %s
			AND l.is_archived = FALSE
			AND l.deleted_at IS NULL 
			AND lt.deleted_at IS NULL 
			AND lt.is_archived = FALSE
		GROUP BY l.location_id
	`

	if params.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", params.Limit)
	}

	args := []interface{}{
		params.UserID,
		params.PermissionNames,
		params.Name,
	}
	condition := ""
	if len(params.LocationIDs) > 0 {
		condition += "and l.location_id = ANY($4) "
		args = append(args, params.LocationIDs)
	}
	query = fmt.Sprintf(query, condition)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locationIDs = make([]string, 0)
	for rows.Next() {
		locationID := &pgtype.Text{}
		fields := []interface{}{locationID}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}
		locationIDs = append(locationIDs, locationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locationIDs, nil
}

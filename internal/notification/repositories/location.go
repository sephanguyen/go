package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LocationRepo struct{}

func (repo *LocationRepo) GetLowestGrantedLocationsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, userID string, permissions []string) ([]string, map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetOrgLocationByUserID")
	defer span.End()

	query := `
		SELECT l1.location_id, l1.access_path
		FROM user_group_member ugm
			JOIN user_group ug ON ugm.user_group_id = ug.user_group_id
			JOIN granted_role gr ON ug.user_group_id = gr.user_group_id
			JOIN role r ON gr.role_id = r.role_id
			JOIN permission_role pr ON r.role_id = pr.role_id
			JOIN permission p ON p.permission_id = pr.permission_id
			JOIN granted_role_access_path grap ON gr.granted_role_id = grap.granted_role_id
			JOIN locations l ON l.location_id = grap.location_id
			JOIN locations l1 ON l1.access_path ~ l.access_path AND l.resource_path = l1.resource_path
		WHERE ugm.user_id = $1::TEXT
			AND p.permission_name = ANY($2::TEXT[])
			AND ugm.deleted_at IS NULL 
			AND ug.deleted_at IS NULL 
			AND gr.deleted_at IS NULL 
			AND r.deleted_at IS NULL 
			AND pr.deleted_at IS NULL 
			AND p.deleted_at IS NULL 
			AND grap.deleted_at IS NULL 
			AND l.deleted_at IS NULL
			AND l1.deleted_at IS NULL
		GROUP BY l1.location_id, l1.access_path;
	`

	rows, err := db.Query(ctx, query, userID, permissions)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	mapLocationIDAndAccessPath := make(map[string]string)
	locationIDs := make([]string, 0)
	for rows.Next() {
		locationID := &pgtype.Text{}
		accessPath := &pgtype.Text{}
		fields := []interface{}{locationID, accessPath}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, nil, err
		}
		mapLocationIDAndAccessPath[locationID.String] = accessPath.String
		locationIDs = append(locationIDs, locationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(locationIDs) == 0 {
		return nil, nil, fmt.Errorf("lowest granted location not found for user: %s, permission: %+v", userID, permissions)
	}

	return locationIDs, mapLocationIDAndAccessPath, nil
}

func (repo *LocationRepo) GetGrantedLocationsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, userID string, permissions []string) ([]string, map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetOrgLocationByUserID")
	defer span.End()

	query := `
		SELECT l.location_id, l.access_path
		FROM user_group_member ugm
			JOIN user_group ug ON ugm.user_group_id = ug.user_group_id
			JOIN granted_role gr ON ug.user_group_id = gr.user_group_id
			JOIN role r ON gr.role_id = r.role_id
			JOIN permission_role pr ON r.role_id = pr.role_id
			JOIN permission p ON p.permission_id = pr.permission_id
			JOIN granted_role_access_path grap ON gr.granted_role_id = grap.granted_role_id
			JOIN locations l ON l.location_id = grap.location_id
		WHERE ugm.user_id = $1::TEXT
			AND p.permission_name = ANY($2::TEXT[])
			AND ugm.deleted_at IS NULL 
			AND ug.deleted_at IS NULL 
			AND gr.deleted_at IS NULL 
			AND r.deleted_at IS NULL 
			AND pr.deleted_at IS NULL 
			AND p.deleted_at IS NULL 
			AND grap.deleted_at IS NULL 
			AND l.deleted_at IS NULL
		GROUP BY l.location_id, l.access_path;
	`

	rows, err := db.Query(ctx, query, userID, permissions)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	mapLocationIDAndAccessPath := make(map[string]string)
	locationIDs := make([]string, 0)
	for rows.Next() {
		locationID := &pgtype.Text{}
		accessPath := &pgtype.Text{}
		fields := []interface{}{locationID, accessPath}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, nil, err
		}
		mapLocationIDAndAccessPath[locationID.String] = accessPath.String
		locationIDs = append(locationIDs, locationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(locationIDs) == 0 {
		return nil, nil, fmt.Errorf("granted location not found for user: %s, permission: %+v", userID, permissions)
	}

	return locationIDs, mapLocationIDAndAccessPath, nil
}

// Get location and access path by location_id, return map[location_id]access_path
func (repo *LocationRepo) GetLocationAccessPathsByIDs(ctx context.Context, db database.QueryExecer, locationIDs []string) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetOrgLocationByUserID")
	defer span.End()

	query := `
		SELECT l.location_id, l.access_path
		FROM locations l
		WHERE l.location_id = ANY($1) AND l.deleted_at IS NULL
	`

	rows, err := db.Query(ctx, query, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapLocationIDAndAccessPath := make(map[string]string)
	for rows.Next() {
		locationID := &pgtype.Text{}
		accessPath := &pgtype.Text{}
		fields := []interface{}{locationID, accessPath}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}

		mapLocationIDAndAccessPath[locationID.String] = accessPath.String
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mapLocationIDAndAccessPath, nil
}

// Get location_ids by location_ids
func (repo *LocationRepo) GetLowestLocationIDsByIDs(ctx context.Context, db database.QueryExecer, locationIDs []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLowestLocationIDsByIDs")
	defer span.End()

	query := `
		SELECT l1.location_id
		FROM locations l
			JOIN locations l1 ON l1.access_path ~ l.access_path AND l.resource_path = l1.resource_path
		WHERE l.location_id = ANY($1) 
			AND l.deleted_at IS NULL
			AND l1.deleted_at IS NULL;
	`
	rows, err := db.Query(ctx, query, locationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locationIDsReturn := make([]string, 0)
	for rows.Next() {
		locationID := &pgtype.Text{}
		fields := []interface{}{locationID}
		err := rows.Scan(fields...)
		if err != nil {
			return nil, err
		}

		locationIDsReturn = append(locationIDsReturn, locationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locationIDsReturn, nil
}

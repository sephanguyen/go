package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type LocationRepo struct{}

func (l *LocationRepo) getLocationByID(ctx context.Context, db database.Ext, id string) (*Location, error) {
	location := &Location{}
	fields, values := location.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE location_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	err := db.QueryRow(ctx, query, &id).Scan(values...)
	return location, err
}

func (l *LocationRepo) GetLocationByID(ctx context.Context, db database.Ext, id string) (*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByID")
	defer span.End()

	location, err := l.getLocationByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return location.ToLocationEntity(), nil
}

func (l *LocationRepo) GetLocationByPartnerInternalID(ctx context.Context, db database.Ext, id string) (*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByPartnerInternalID")
	defer span.End()

	location := &Location{}
	fields, values := location.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM locations
		WHERE partner_internal_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	if err := db.QueryRow(ctx, query, &id).Scan(values...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return location.ToLocationEntity(), nil
}

func (l *LocationRepo) UpsertLocations(ctx context.Context, db database.Ext, locations []*domain.Location) error {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.UpsertLocations")
	defer span.End()
	b := &pgx.Batch{}
	bUpdate := &pgx.Batch{}
	locationIDs := sliceutils.Map(locations, func(l *domain.Location) string {
		return l.LocationID
	})
	existLocations, err := l.GetLocationsByLocationIDs(ctx, db, database.TextArray(locationIDs), false)
	if err != nil {
		return err
	}
	existIDs := sliceutils.Map(existLocations, func(l *domain.Location) string {
		return l.LocationID
	})
	for order, location := range locations {
		locationDto, _ := NewLocationFromEntity(location)
		// This is for ordering the added locations
		locationDto.UpdatedAt = database.Timestamptz(time.Now().Add(time.Second * time.Duration(order)))
		locationDto.CreatedAt = locationDto.UpdatedAt
		fields, args := locationDto.FieldMapWithoutRP()
		placeHolders := database.GeneratePlaceholders(len(fields))

		if !slices.Contains(
			existIDs, location.LocationID) {
			query := fmt.Sprintf("INSERT INTO locations (%s) VALUES (%s)", strings.Join(fields, ", "), placeHolders)
			b.Queue(query, args...)
		} else {
			query := fmt.Sprintf(`UPDATE locations SET updated_at = $6,
			name = $1, parent_location_id = $2, partner_internal_parent_id = $3, deleted_at = NULL, is_archived = $4, access_path = $5 where location_id='%s'`, location.LocationID)
			bUpdate.Queue(query, locationDto.Name, locationDto.ParentLocationID, locationDto.PartnerInternalParentID, locationDto.IsArchived, locationDto.AccessPath, locationDto.UpdatedAt)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return err
		}
	}
	result = db.SendBatch(ctx, bUpdate)
	defer result.Close()
	for i := 0; i < bUpdate.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func convertErrToErrResForEachLineCSV(err error, i int, method string) *domain.UpsertErrors {
	return &domain.UpsertErrors{
		RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
		Error:     fmt.Sprintf("unable to %s location item: %s", method, err),
	}
}

func (l *LocationRepo) DeleteByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.DeleteByPartnerInternalIDs")
	defer span.End()

	query := "UPDATE locations SET deleted_at = now(), updated_at = now() WHERE partner_internal_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &ids)
	if err != nil {
		return err
	}

	return nil
}

func (l *LocationRepo) GetLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationsByPartnerInternalIDs")
	defer span.End()
	t := &Location{}

	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE partner_internal_id = ANY ($1) AND deleted_at IS NULL", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*domain.Location
	for rows.Next() {
		p := new(Location)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (l *LocationRepo) GetLocationsByLocationIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationsByLocationIDs")
	defer span.End()
	t := &Location{}
	condition := "AND deleted_at IS NULL"
	if allowDeleted {
		condition = ""
	}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE location_id = ANY ($1) %s", strings.Join(fields, ","), t.TableName(), condition)
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ls []*domain.Location
	for rows.Next() {
		l := new(Location)
		if err := rows.Scan(database.GetScanFields(l, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ls = append(ls, l.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return ls, nil
}

func (l *LocationRepo) RetrieveLocations(ctx context.Context, db database.Ext, queries domain.FilterLocation) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.RetrieveLocations")
	defer span.End()
	fields := database.GetFieldNames(&Location{})
	filterIsArchived := ""
	if !queries.IncludeIsArchived {
		filterIsArchived = "is_archived = true OR "
	}
	query := fmt.Sprintf(`SELECT %s FROM locations 
						  WHERE location_id not in ( WITH RECURSIVE al AS ( SELECT location_id FROM locations
													WHERE %s deleted_at IS NOT NULL
													UNION 
														SELECT l.location_id 
														FROM locations AS l 
														JOIN al ON al.location_id = l.parent_location_id) SELECT * FROM al )
						   ORDER BY updated_at, created_at`, strings.Join(fields, ","), filterIsArchived)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	return scanLocations(rows, fields)
}

func (l *LocationRepo) GetLocationByLocationTypeName(ctx context.Context, db database.Ext, name string) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByLocationTypeName")
	defer span.End()
	fields := database.GetFieldNames(&Location{})
	query := `SELECT l.location_id, l.name, l.location_type, l.parent_location_id, l.partner_internal_id, l.partner_internal_parent_id, l.is_archived, l.updated_at, l.created_at, l.deleted_at, l.access_path, l.resource_path
		from locations l join location_types lt ON lt.location_type_id = l.location_type 
		where lt.name = $1 and l.deleted_at is null and lt.deleted_at is null`

	rows, err := db.Query(ctx, query, &name)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	return scanLocations(rows, fields)
}

func (l *LocationRepo) GetLocationByLocationTypeID(ctx context.Context, db database.Ext, id string) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByLocationTypeID")
	defer span.End()
	fields := database.GetFieldNames(&Location{})
	query := `
		SELECT location_id, name, location_type, parent_location_id, partner_internal_id, partner_internal_parent_id, is_archived, updated_at, created_at, deleted_at, access_path, resource_path
		from locations 
		where location_type = $1 and deleted_at is null`

	rows, err := db.Query(ctx, query, &id)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	return scanLocations(rows, fields)
}

func (l *LocationRepo) UpdateAccessPath(ctx context.Context, db database.Ext, ids []string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.UpdateAccessPath")
	defer span.End()
	query := fmt.Sprintf(`WITH RECURSIVE with_locations(location_id, name, parent_location_id, access_path) AS (
			SELECT l.location_id , l.name , l.parent_location_id , l.location_id::TEXT AS access_path 
			FROM locations AS l 
			WHERE l.parent_location_id IS NULL
		UNION ALL
			SELECT lo.location_id, lo.name, lo.parent_location_id, (wl.access_path || '/' || lo.location_id::TEXT) 
			FROM with_locations AS wl, locations AS lo 
			WHERE lo.parent_location_id = wl.location_id
		) UPDATE locations 
		set access_path = with_locations.access_path
		from with_locations
		where with_locations.location_id = locations.location_id and locations.location_id = ANY ($1);`,
	)
	if _, err := db.Exec(ctx, query, &ids); err != nil {
		return err
	}

	return nil
}

func (l *LocationRepo) GetLocationOrg(ctx context.Context, db database.Ext, resourcePath string) (*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationOrg")
	defer span.End()

	fields := database.GetFieldNames(&Location{})
	query := fmt.Sprintf(
		`
			SELECT l.%s
			FROM locations l
			INNER JOIN location_types lt ON l.location_type = lt.location_type_id
			WHERE l.deleted_at IS NULL
				AND lt.name = $1
				AND l.resource_path = $2
				AND (l.parent_location_id is null or (length(l.parent_location_id) = 0))
		`, strings.Join(fields, ", l."))

	rows, err := db.Query(ctx, query, domain.DefaultLocationType, resourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	location := &Location{}
	for rows.Next() {
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return location.ToLocationEntity(), nil
}

type GetLowestLevelLocationsParams struct {
	Name        string
	Limit       int32
	Offset      int32
	LocationIDs []string
}

// Get locations belongs to location type that has lowest level (highest number)
func (l *LocationRepo) GetLowestLevelLocationsV2(ctx context.Context, db database.Ext, params *GetLowestLevelLocationsParams) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLowestLevelLocationsV2")
	defer span.End()
	rawQuery := `SELECT l.location_id, l.name
	FROM locations l 
	JOIN location_types lt ON l.location_type  = lt.location_type_id 
	WHERE l.deleted_at IS NULL 
	AND l.is_archived = FALSE
	AND lt.deleted_at IS NULL 
	AND lt.is_archived = FALSE
	AND lt.level = (SELECT MAX(level) FROM location_types lt2 
		WHERE lt2.deleted_at IS NULL AND lt2.is_archived = FALSE)
	AND lower(l.name) like lower(CONCAT('%%',$1::text,'%%')) 
	%s`

	cond := ""
	args := []interface{}{
		params.Name,
	}
	if len(params.LocationIDs) > 0 {
		cond += "and l.location_id = ANY($2) "
		args = append(args, params.LocationIDs)
	}

	q := fmt.Sprintf(rawQuery, cond)

	if params.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", params.Limit)
	}
	if params.Offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", params.Offset)
	}
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()
	locs := make([]*domain.Location, 0)
	for rows.Next() {
		l := &Location{}
		if err := rows.Scan(database.GetScanFields(l, []string{"location_id", "name"})...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locs = append(locs, l.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locs, nil
}

func (l *LocationRepo) RetrieveLowestLevelLocations(ctx context.Context, db database.Ext, params *GetLowestLevelLocationsParams) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.RetrieveLowestLevelLocations")
	defer span.End()
	rawQuery := `select location_id,name from locations 
	where location_type not in (select distinct parent_location_type_id from location_types where parent_location_type_id is not null 
		                        and deleted_at is null and is_archived = false)
	and  lower(name) like lower(CONCAT('%%',$1::text,'%%')) and deleted_at is null and is_archived = false %s
	order by created_at desc`
	// condition
	condition := ""
	args := []interface{}{
		params.Name,
	}
	if len(params.LocationIDs) > 0 {
		condition += "and location_id = ANY($2) "
		args = append(args, params.LocationIDs)
	}

	query := fmt.Sprintf(rawQuery, condition)

	if params.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", params.Limit)
	}
	if params.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", params.Offset)
	}
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	locations := make([]*domain.Location, 0)
	for rows.Next() {
		location := &Location{}
		if err := rows.Scan(database.GetScanFields(location, []string{"location_id", "name"})...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locations = append(locations, location.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locations, nil
}

// GetAllLocations
// Location.LocationType is the name of location
// Really confusing name of LocationType column, so please choose the right thing you need
func (l *LocationRepo) GetAllLocations(ctx context.Context, db database.Ext) ([]*Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetAllLocations")
	defer span.End()
	fields := database.GetFieldNames(&Location{})

	// Join to map location_type.name as location_type
	remapped := sliceutils.Map(fields, func(s string) string {
		if s == "location_type" {
			return "SQ.location_type_name as location_type"
		}
		return "SQ." + s
	})
	innerMapped := sliceutils.Map(fields, func(s string) string {
		return "L1." + s
	})
	query := fmt.Sprintf(`SELECT %s FROM 
						(SELECT %s, lt."name" as location_type_name FROM locations L1
						join location_types lt 
						on L1.location_type  = lt.location_type_id 
						  WHERE location_id not in ( WITH RECURSIVE al AS ( SELECT location_id FROM locations
													WHERE deleted_at IS NOT NULL
													UNION 
														SELECT l.location_id 
														FROM locations AS l 
														JOIN al ON al.location_id = l.parent_location_id) SELECT * FROM al )
						   ORDER BY updated_at, created_at) as SQ
	`, strings.Join(remapped, ","), strings.Join(innerMapped, ","))

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	locations := []*Location{}
	for rows.Next() {
		location := &Location{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locations, nil
}

// GetAllRawLocations
// Location.LocationType is the foreign key
// Really confusing name of LocationType column, so please choose the right thing you need
func (l *LocationRepo) GetAllRawLocations(ctx context.Context, db database.Ext) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetAllRawLocations")
	defer span.End()
	e := &Location{}
	fields := database.GetFieldNames(e)

	parentCond := `location_id not in ( 
						WITH RECURSIVE al AS ( 
							SELECT location_id FROM locations
								WHERE deleted_at IS NOT NULL
								UNION 
								SELECT l.location_id 
								FROM locations AS l 
								JOIN al ON al.location_id = l.parent_location_id
						) SELECT * FROM al 
						)`
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY updated_at, created_at;`,
		strings.Join(fields, ","),
		e.TableName(),
		parentCond,
	)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	return scanLocations(rows, fields)
}

func scanLocations(rows pgx.Rows, fields []string) ([]*domain.Location, error) {
	defer rows.Close()
	var locations []*domain.Location
	for rows.Next() {
		location := &Location{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locations = append(locations, location.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locations, nil
}

func (l *LocationRepo) GetLocationByLocationTypeIDs(ctx context.Context, db database.Ext, ids []string) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetLocationByLocationTypeIDs")
	defer span.End()
	table := &Location{}
	fields := database.GetFieldNames(table)
	query := fmt.Sprintf(`
	SELECT %s
	from locations 
	where location_type = ANY($1) and deleted_at is null`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	return scanLocations(rows, fields)
}

func (l *LocationRepo) GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*domain.Location, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetChildLocations")
	defer span.End()
	table := &Location{}
	fields := database.GetFieldNames(table)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE access_path like CONCAT('%%',$1::text,'%%')", strings.Join(fields, ","), table.TableName())
	rows, err := db.Query(ctx, query, &id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		location := new(Location)
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locations = append(locations, location.ToLocationEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locations, nil
}

func (l *LocationRepo) GetRootLocation(ctx context.Context, db database.Ext) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationRepo.GetRootLocation")
	defer span.End()
	query := `SELECT l.location_id FROM locations l where parent_location_id is null and deleted_at is null limit 1`

	var locationID string

	err := db.QueryRow(ctx, query).Scan(&locationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", errors.Wrap(err, "db.Query")
	}
	return locationID, nil
}

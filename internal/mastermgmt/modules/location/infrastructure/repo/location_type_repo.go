package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LocationTypeRepo struct{}

func (l *LocationTypeRepo) getLocationTypeByID(ctx context.Context, db database.Ext, id string) (*LocationType, error) {
	locationType := &LocationType{}
	fields, values := locationType.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM location_types
		WHERE location_type_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	if err := db.QueryRow(ctx, query, &id).Scan(values...); err != nil {
		return nil, err
	}

	return locationType, nil
}

func (l *LocationTypeRepo) GetLocationTypeByID(ctx context.Context, db database.Ext, id string) (*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByID")
	defer span.End()

	locationType, err := l.getLocationTypeByID(ctx, db, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return locationType.ToLocationTypeEntity(), nil
}

func (l *LocationTypeRepo) GetLocationTypeByName(ctx context.Context, db database.Ext, name string, allowEmpty bool) (*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByName")
	defer span.End()

	locationType := &LocationType{}
	fields, values := locationType.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM location_types
		WHERE name = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	if err := db.QueryRow(ctx, query, &name).Scan(values...); err != nil {
		if err == pgx.ErrNoRows && allowEmpty {
			return nil, nil
		}
		return nil, err
	}

	return locationType.ToLocationTypeEntity(), nil
}

func (l *LocationTypeRepo) UpsertLocationTypes(ctx context.Context, db database.Ext, locationTypes map[int]*domain.LocationType) (errors []*domain.UpsertErrors) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.UpsertLocationTypes")
	defer span.End()
	b := &pgx.Batch{}
	mappers := make(map[int]int)
	i := 0
	for order, locationType := range locationTypes {
		mappers[i] = order
		i++
		locationTypeDto, _ := NewLocationTypeFromEntity(locationType)
		fields, args := locationTypeDto.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf("INSERT INTO location_types (%s) "+
			"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__location_type_name_resource_path DO "+
			"UPDATE SET updated_at = now(), display_name = $3, parent_name = $4, parent_location_type_id = $5, deleted_at = NULL", strings.Join(fields, ", "), placeHolders)
		b.Queue(query, args...)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			errors = append(errors, convertErrToErrResForEachLineCSV(fmt.Errorf("location type invalid"), mappers[i], "parse"))
			continue
		}
	}
	return errors
}

func (l *LocationTypeRepo) Import(ctx context.Context, db database.Ext, locationTypes []*domain.LocationType) error {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.Import")
	defer span.End()
	b := &pgx.Batch{}
	for _, locationType := range locationTypes {
		locationTypeDto, _ := NewLocationTypeFromEntity(locationType)
		if locationTypeDto.CreatedAt.Time.IsZero() {
			locationTypeDto.CreatedAt = database.Timestamptz(time.Now())
		}
		locationTypeDto.UpdatedAt = database.Timestamptz(time.Now())
		fields, args := locationTypeDto.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO location_types (%s) 
			VALUES (%s) ON CONFLICT ON CONSTRAINT unique__location_type_name_resource_path DO
			UPDATE SET updated_at = now(), display_name = $3, deleted_at = NULL, level = $10, is_archived = $6`,
			strings.Join(fields, ", "),
			placeHolders)
		b.Queue(query, args...)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("location types could not be imported")
		}
	}
	return nil
}

func (l *LocationTypeRepo) GetLocationTypeByNames(ctx context.Context, db database.Ext, names pgtype.TextArray) ([]*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByNames")
	defer span.End()
	t := &LocationType{}

	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE name = ANY ($1) AND deleted_at IS NULL ", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*domain.LocationType
	for rows.Next() {
		p := new(LocationType)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p.ToLocationTypeEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (l *LocationTypeRepo) GetLocationTypeByIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByIDs")
	defer span.End()
	t := &LocationType{}
	condition := "AND deleted_at IS NULL"
	if allowDeleted {
		condition = ""
	}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE location_type_id = ANY ($1) %s", strings.Join(fields, ","), t.TableName(), condition)
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*domain.LocationType
	for rows.Next() {
		p := new(LocationType)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p.ToLocationTypeEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (l *LocationTypeRepo) DeleteByPartnerNames(ctx context.Context, db database.Ext, names pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.DeleteByPartnerNames")
	defer span.End()

	query := "UPDATE location_types SET deleted_at = now(), updated_at = now() WHERE name = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &names)
	if err != nil {
		return err
	}

	return nil
}

func (l *LocationTypeRepo) RetrieveLocationTypes(ctx context.Context, db database.Ext) ([]*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.RetrieveLocationTypes")
	defer span.End()

	locTypeDtos, err := l.GetAllLocationTypes(ctx, db)
	if err != nil {
		return nil, err
	}
	locationTypes := []*domain.LocationType{}
	for _, locDto := range locTypeDtos {
		locationTypes = append(locationTypes, locDto.ToLocationTypeEntity())
	}
	return locationTypes, nil
}

func (l *LocationTypeRepo) RetrieveLocationTypesV2(ctx context.Context, db database.Ext) ([]*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.RetrieveLocationTypesV2")
	defer span.End()

	locTypeDtos, err := l.GetAllLocationTypesV2(ctx, db)
	if err != nil {
		return nil, err
	}
	locTypes := []*domain.LocationType{}
	for _, locDto := range locTypeDtos {
		locTypes = append(locTypes, locDto.ToLocationTypeEntity())
	}
	return locTypes, nil
}

func (l *LocationTypeRepo) GetLocationTypeByParentName(ctx context.Context, db database.Ext, parentName string) (*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByParentName")
	defer span.End()
	locationType := &LocationType{}
	fields, values := locationType.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE parent_name = $1 AND deleted_at IS NULL ", strings.Join(fields, ","), locationType.TableName())
	err := db.QueryRow(ctx, query, &parentName).Scan(values...)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return locationType.ToLocationTypeEntity(), nil
}

func (l *LocationTypeRepo) GetLocationTypeByNameAndParent(ctx context.Context, db database.Ext, name, parentName string) (*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByNameAndParent")
	defer span.End()

	locationType := &LocationType{}
	fields, values := locationType.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM location_types
		WHERE name = $1 and parent_name = $2
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	if err := db.QueryRow(ctx, query, &name, &parentName).Scan(values...); err != nil {
		return nil, err
	}

	return locationType.ToLocationTypeEntity(), nil
}

func (l *LocationTypeRepo) GetAllLocationTypes(ctx context.Context, db database.Ext) ([]*LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetAllLocationTypes")
	defer span.End()
	fields := database.GetFieldNames(&LocationType{})
	query := fmt.Sprintf(`SELECT %s FROM location_types 
						WHERE location_type_id not in ( WITH RECURSIVE alt AS (
							SELECT location_type_id 
   								FROM location_types
   								WHERE is_archived = true or deleted_at is not null  
   								UNION 
   									SELECT lt.location_type_id 
  		 							FROM location_types AS lt
           							JOIN alt ON alt.location_type_id = lt.parent_location_type_id) SELECT * FROM alt) 
						ORDER BY level ASC`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	locationTypes := []*LocationType{}
	for rows.Next() {
		location := &LocationType{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locationTypes = append(locationTypes, location)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locationTypes, nil
}

func (l *LocationTypeRepo) GetAllLocationTypesV2(ctx context.Context, db database.Ext) ([]*LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetAllLocationTypesV2")
	defer span.End()
	fields := database.GetFieldNames(&LocationType{})
	query := fmt.Sprintf(`SELECT %s FROM location_types 
						WHERE is_archived = false AND deleted_at is null  
						ORDER BY parent_name ASC`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	locTypes := []*LocationType{}
	for rows.Next() {
		location := &LocationType{}
		if err := rows.Scan(database.GetScanFields(location, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locTypes = append(locTypes, location)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locTypes, nil
}

func (l *LocationTypeRepo) UpdateLevels(ctx context.Context, db database.Ext) error {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.UpdateLevels")
	defer span.End()

	query := `WITH RECURSIVE type_dept (id, dept) AS (
		SELECT location_type_id, 0
		FROM location_types
		WHERE parent_location_type_id  IS null
		UNION ALL
		SELECT lt.location_type_id , d.dept + 1
		FROM location_types lt
		JOIN type_dept d ON lt.parent_location_type_id  = d.id
	  )
	  UPDATE location_types 
	  SET "level" = type_dept.dept
	  FROM type_dept
	  WHERE type_dept.id = location_type_id`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (l *LocationTypeRepo) GetLocationTypesByLevel(ctx context.Context, db database.Ext, level string) ([]*domain.LocationType, error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationTypeRepo.GetLocationTypeByIDs")
	defer span.End()
	t := &LocationType{}

	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE level = $1 AND deleted_at IS NULL", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locTypes []*domain.LocationType
	for rows.Next() {
		locType := new(LocationType)
		if err := rows.Scan(database.GetScanFields(locType, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locTypes = append(locTypes, locType.ToLocationTypeEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return locTypes, nil
}

package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/pkg/errors"
)

type DomainLocationRepo struct{}

type LocationAttribute struct {
	ID                field.String
	Name              field.String
	IsArchived        field.Boolean
	PartnerInternalID field.String
	OrganizationID    field.String
}

type Location struct {
	LocationAttribute

	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func NewLocation(location entity.DomainLocation) *Location {
	now := field.NewTime(time.Now())
	return &Location{
		LocationAttribute: LocationAttribute{
			ID:                location.LocationID(),
			Name:              location.Name(),
			IsArchived:        location.IsArchived(),
			PartnerInternalID: location.PartnerInternalID(),
			OrganizationID:    location.OrganizationID(),
		},
		UpdatedAt: now,
		CreatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (location *Location) LocationID() field.String {
	return location.ID
}
func (location *Location) Name() field.String {
	return location.LocationAttribute.Name
}
func (location *Location) IsArchived() field.Boolean {
	return location.LocationAttribute.IsArchived
}
func (location *Location) PartnerInternalID() field.String {
	return location.LocationAttribute.PartnerInternalID
}
func (location *Location) OrganizationID() field.String {
	return location.LocationAttribute.OrganizationID
}

func (location *Location) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id",
			"name",
			"is_archived",
			"partner_internal_id",
			"resource_path",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&location.ID,
			&location.LocationAttribute.Name,
			&location.LocationAttribute.IsArchived,
			&location.LocationAttribute.PartnerInternalID,
			&location.LocationAttribute.OrganizationID,
			&location.UpdatedAt,
			&location.CreatedAt,
			&location.DeletedAt,
		}
}

func (location *Location) TableName() string {
	return "locations"
}

func (r *DomainLocationRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.DomainLocations, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainLocationRepo.GetByPartnerInternalIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE partner_internal_id = ANY($1) and deleted_at is NULL`
	location := NewLocation(entity.NullDomainLocation{})

	fieldNames, _ := location.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		location.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(partnerInternalIDs),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entity.DomainLocation
	for rows.Next() {
		item := NewLocation(entity.NullDomainLocation{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainLocationRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) (entity.DomainLocations, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainLocationRepo.GetByIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE location_id = ANY($1) and deleted_at is NULL`
	location := NewLocation(entity.NullDomainLocation{})

	fieldNames, _ := location.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		location.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(ids),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entity.DomainLocation
	for rows.Next() {
		item := NewLocation(entity.NullDomainLocation{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}

func (r *DomainLocationRepo) RetrieveLowestLevelLocations(ctx context.Context, db database.Ext, name string, limit int32, offset int32, locationIDs []string) (entity.DomainLocations, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainLocationRepo.RetrieveLowestLevelLocations")
	defer span.End()
	statement := `
		SELECT %s
		FROM %s
		WHERE
			location_type NOT IN (
				SELECT distinct parent_location_type_id
				FROM location_types
				WHERE
					parent_location_type_id IS NOT NULL AND
					deleted_at IS NULL AND is_archived = false
			) AND
			LOWER(name) LIKE LOWER(CONCAT('%%',$1::text,'%%')) AND
			deleted_at IS NULL AND
			is_archived = FALSE %s
		ORDER BY created_at DESC
	`
	// condition
	condition := ""
	args := []interface{}{name}
	if len(locationIDs) > 0 {
		condition += "and location_id = ANY($2) "
		args = append(args, locationIDs)
	}

	location := NewLocation(entity.NullDomainLocation{})
	fieldNames, _ := location.FieldMap()

	query := fmt.Sprintf(
		statement,
		strings.Join(fieldNames, ","),
		location.TableName(),
		condition,
	)

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}
	defer rows.Close()

	locations := make(entity.DomainLocations, 0, len(locationIDs))
	for rows.Next() {
		location := NewLocation(entity.NullDomainLocation{})

		_, fieldValues := location.FieldMap()

		err := rows.Scan(fieldValues...)

		if err != nil {
			return nil, InternalError{
				RawError: fmt.Errorf("row.Scan: %w", err),
			}
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "rows.Err"),
		}
	}
	return locations, nil
}

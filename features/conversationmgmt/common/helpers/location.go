package helpers

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (helper *ConversationMgmtHelper) CreateLocationWithDB(ctx context.Context, resourcePath string, typeLocName string, parentLocID string, parentLocTypeID string) (*entities.Location, error) {
	locationID := idutil.ULIDNow()
	newLocation := &entities.Location{
		ID:               locationID,
		Name:             typeLocName + "-" + locationID,
		AccessPath:       locationID,
		ParentLocationID: parentLocID,
		TypeLocation:     typeLocName,
		TypeLocationID:   "",
	}

	parentLocation := pgtype.Text{Status: pgtype.Null}
	locationTypeID := pgtype.Text{Status: pgtype.Null}

	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})
	if parentLocID != "" {
		var parentAp string
		query := `
			SELECT access_path FROM locations WHERE location_id = $1
		`
		if err := helper.BobPostgresDBConn.QueryRow(ctx2, query, parentLocID).Scan(&parentAp); err != nil {
			return nil, fmt.Errorf("cannot get parent location %s: %v", parentLocID, err)
		}
		newLocation.AccessPath = fmt.Sprintf("%s/%s", parentAp, locationID)
		parentLocation = database.Text(parentLocID)
	}

	if typeLocName != "" {
		locTypeID, err := helper.createLocationTypeWithDB(ctx, resourcePath, typeLocName, parentLocTypeID)
		if err != nil {
			return nil, fmt.Errorf("helper.createLocationTypeWithDB: %v", err)
		}
		locationTypeID = database.Text(locTypeID)
		newLocation.TypeLocationID = locationTypeID.String
	}

	stmt := `
		INSERT INTO public.locations
			(location_id, name, location_type, parent_location_id, updated_at, created_at, access_path)
		VALUES ($1, $2, $3, $4, now(), now(), $5)
	`
	if _, err := helper.BobPostgresDBConn.Exec(ctx2, stmt, locationID, newLocation.Name, locationTypeID, parentLocation, newLocation.AccessPath); err != nil {
		return nil, err
	}

	return newLocation, nil
}

func (helper *ConversationMgmtHelper) createLocationTypeWithDB(ctx context.Context, resourcePath string, typeLocName string, parentLocTypeID string) (string, error) {
	parentLocationTypeID := pgtype.Text{Status: pgtype.Null}
	parentLocationTypeName := pgtype.Text{Status: pgtype.Null}
	locTypeID := ""

	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	if typeLocName == "" {
		return "", fmt.Errorf("location type name is missing")
	}

	querySelectLocType := "SELECT location_type_id FROM location_types WHERE name = $1"
	if err := helper.BobDBConn.QueryRow(ctx2, querySelectLocType, database.Text(typeLocName)).Scan(&locTypeID); err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	if parentLocTypeID != "" {
		var parentLocTypeName string
		query := `
			SELECT location_type_id, name FROM location_types WHERE location_type_id = $1
		`
		if err := helper.BobDBConn.QueryRow(ctx2, query, parentLocTypeID).Scan(&parentLocTypeID, &parentLocTypeName); err != nil {
			return "", fmt.Errorf("cannot get parent location type %s: %v", parentLocTypeID, err)
		}
		parentLocationTypeID = database.Text(parentLocTypeID)
		parentLocationTypeName = database.Text(parentLocTypeName)
	}

	if locTypeID == "" {
		locTypeID = idutil.ULIDNow()
		stmt := `
				INSERT INTO public.location_types
					(location_type_id, name, display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived)
				VALUES($1, $2, $3, $4, $5, now(), now(), NULL, autofillresourcepath(), false);
			`
		if _, err := helper.BobDBConn.Exec(ctx2, stmt, locTypeID, typeLocName, typeLocName, parentLocationTypeName, parentLocationTypeID); err != nil {
			return "", err
		}
	}

	return locTypeID, nil
}

package common

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// This is for document purpurse, you can always use concrete struct instead of this interface
// Ctx should have token of school admin already
type LocationSuite interface {
	// CreateLocations create org location for the given resource path, then create more totalLoc children of that location
	// Use this when MasterMgmt API is mature enough
	CreateLocationWithAPI(ctx context.Context, rp string, totalLoc int) ([]string, error)

	// CreateLocations using direct db insert
	// if parentLoc is empty, created location is org location
	CreateLocationWithDB(ctx context.Context, rp string, parentLoc string) (string, error)
}

func (s *suite) CreateLocationWithDB(ctx context.Context, resourcePath string, typeLocName string, parentLocID string, parentLocTypeID string) (string, string, error) {
	locationID := idutil.ULIDNow()
	parentLocation := pgtype.Text{Status: pgtype.Null}
	locationTypeID := pgtype.Text{Status: pgtype.Null}
	accessPath := locationID
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})
	if parentLocID != "" {
		var parentAp string
		query := `
			SELECT access_path FROM locations WHERE location_id=$1
		`

		if err := s.BobPostgresDB.QueryRow(ctx2, query, parentLocID).Scan(&parentAp); err != nil {
			return "", "", err
		}
		accessPath = fmt.Sprintf("%s/%s", parentAp, locationID)
		parentLocation = database.Text(parentLocID)
	}

	if typeLocName != "" {
		locTypeID, err := s.createLocationTypeWithDB(ctx, resourcePath, typeLocName, parentLocTypeID)
		if err != nil {
			return "", "", fmt.Errorf("helper.createLocationTypeWithDB: %v", err)
		}
		locationTypeID = database.Text(locTypeID)
	}

	stmt := `
		INSERT INTO public.locations
		(location_id,name, location_type, parent_location_id, updated_at, created_at,access_path)
		VALUES ($1, $1, $2, $3, now(), now(), $4)
	`
	if _, err := s.BobPostgresDB.Exec(ctx2, stmt, locationID, locationTypeID, parentLocation, accessPath); err != nil {
		return "", "", err
	}
	return locationID, locationTypeID.String, nil
}

func (s *suite) createLocationTypeWithDB(ctx context.Context, resourcePath string, typeLocName string, parentLocTypeID string) (string, error) {
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
	if err := s.BobDB.QueryRow(ctx2, querySelectLocType, database.Text(typeLocName)).Scan(&locTypeID); err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	if parentLocTypeID != "" {
		var parentLocTypeName string
		query := `
			SELECT location_type_id, name FROM location_types WHERE location_type_id = $1
		`
		if err := s.BobDB.QueryRow(ctx2, query, parentLocTypeID).Scan(&parentLocTypeID, &parentLocTypeName); err != nil {
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
		if _, err := s.BobDB.Exec(ctx2, stmt, locTypeID, typeLocName, typeLocName, parentLocationTypeName, parentLocationTypeID); err != nil {
			return "", err
		}
	}

	return locTypeID, nil
}

// Ctx should already have token metadata
func (s *suite) CreateLocations(ctx context.Context, rp string, totalLoc int) ([]string, error) {
	orgLocationID := idutil.ULIDNow()
	err := s.insertOrgLocationTypesWithRp(ctx, rp)
	if err != nil {
		return nil, err
	}
	err = s.insertOrgLocationWithResourcePath(ctx, rp, orgLocationID)
	if err != nil {
		return nil, err
	}

	locationTypeHeaders := "name,display_name,parent_name,is_archived"
	uniqueLocationType := "brand,Brand,,0"
	importLocTypeReq := &bpb.ImportLocationTypeRequest{
		Payload: []byte(fmt.Sprintf(`%s
			%s`, locationTypeHeaders, uniqueLocationType)),
	}
	uniqueLocationTypeName := "brand"

	res, err := bpb.NewMasterDataImporterServiceClient(s.BobConn).
		ImportLocationType(s.SignedCtx(ctx), importLocTypeReq)
	if err != nil {
		return nil, fmt.Errorf("ImportLocationType %w", err)
	}
	if res.TotalSuccess != 1 {
		return nil, fmt.Errorf("importing location type failed some how, %v", res.Errors)
	}

	// Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id,is_archived
	locationHeaders := "partner_internal_id,name,location_type,partner_internal_parent_id,is_archived"

	// if parent_name == '', it is direct child of org location
	locNames := []string{}
	var locLines = "\n"
	for i := 0; i < totalLoc; i++ {
		locName := idutil.ULIDNow()
		locLines = fmt.Sprintf("%s\n%s,%s name,%s,,0", locLines, locName, locName, uniqueLocationTypeName)
		locNames = append(locNames, locName+" name")
	}
	payload := fmt.Sprintf("%s%s", locationHeaders, locLines)
	fmt.Printf("debug payload\n%s\n", payload)

	importLocReq := &bpb.ImportLocationRequest{
		Payload: []byte(payload),
	}
	locres, err := bpb.NewMasterDataImporterServiceClient(s.BobConn).
		ImportLocation(s.SignedCtx(ctx), importLocReq)
	if err != nil {
		return nil, err
	}
	if locres.TotalSuccess != 1 {
		return nil, fmt.Errorf("importing location failed some how, %v", res.Errors)
	}
	rows, err := s.BobDB.Query(ctx, "select location_id FROM locations where name=ANY($1) and deleted_at IS NULL", database.TextArray(locNames))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make([]string, 0, len(locNames))
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, id)
	}
	if len(ret) != len(locNames) {
		return nil, fmt.Errorf("cannot find enough location with names %v, only %v returned", locNames, ret)
	}
	return ret, nil
}

func (s *suite) insertOrgLocationWithResourcePath(ctx context.Context, rp string, orgLocationID string) error {
	stepState := StepStateFromContext(ctx)

	name := database.Text("org")
	stmt := `INSERT INTO locations (location_id, name, created_at, updated_at, resource_path, location_type, is_archived )
		VALUES ($1, $2, now(), now(), $3, $4, false) ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING`
	_, err := s.BobDB.Exec(ctx, stmt, orgLocationID, name, rp, stepState.LocationTypeOrgID)
	if err != nil {
		return fmt.Errorf("cannot insert location, err: %s", err)
	}

	return nil
}

func (s *suite) insertOrgLocationTypesWithRp(ctx context.Context, rp string) error {
	stepState := StepStateFromContext(ctx)
	randomStr := idutil.ULIDNow()
	name := database.Text("org")
	stmt := `INSERT INTO location_types (location_type_id, name, display_name, resource_path, updated_at, created_at, is_archived)
	VALUES ($1, $2, 'Org', $3, now(), now(), false) ON CONFLICT ON CONSTRAINT unique__location_type_name_resource_path DO NOTHING`
	_, err := s.BobDB.Exec(ctx, stmt, randomStr, name, rp)
	if err != nil {
		return fmt.Errorf("cannot insert location type, err: %s", err)
	}
	var locationTypeID string
	queryLocationTypeOrg := "SELECT location_type_id FROM location_types l WHERE l.name = 'org' AND l.deleted_at IS NULL and l.resource_path = $1"
	if err := s.BobDB.QueryRow(ctx, queryLocationTypeOrg, rp).Scan(&locationTypeID); err != nil {
		return fmt.Errorf("cannot get location type org, err: %s", err)
	}
	stepState.LocationTypeOrgID = locationTypeID
	return nil
}

package mastermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) PrepareLocationWithChildren(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locTypes, err := s.prepareLocationTypes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't prepare location types: %s", err)
	}
	// already ordered by levels
	stepState.LocationTypes = locTypes[:4]

	e := &repo.Location{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf(`
	SELECT %s
	from locations 
	where parent_location_id is null and deleted_at is null`, strings.Join(fields, ", "))
	err = s.BobDBTrace.QueryRow(ctx, query).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cant select org location: %s", err)
	}
	orgLocation := e.ToLocationEntity()
	ids := s.generateULIDs(5)
	treeLocation := &domain.TreeLocation{
		LocationID:        ids[0],
		ParentLocationID:  orgLocation.LocationID,
		PartnerInternalID: fmt.Sprintf("partner_internal_%v", ids[0]),
		LocationType:      locTypes[1].LocationTypeID,
		Name:              "Sub v1 1 " + ids[0],
		AccessPath:        fmt.Sprintf("%s/%s", orgLocation.LocationID, ids[0]), // children 's AP will be calculated
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Children: []*domain.TreeLocation{
			{
				LocationID:        ids[1],
				LocationType:      locTypes[2].LocationTypeID,
				PartnerInternalID: fmt.Sprintf("partner_internal_%v", ids[1]),
				Name:              "Sub v2 1 " + ids[1],
				CreatedAt:         time.Now().AddDate(0, 0, -1), // for ordering
				UpdatedAt:         time.Now(),
			},
			{
				LocationID:        ids[2],
				LocationType:      locTypes[2].LocationTypeID,
				PartnerInternalID: fmt.Sprintf("partner_internal_%v", ids[2]),
				Name:              "Sub v2 2 " + ids[2],
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				Children: []*domain.TreeLocation{
					{
						LocationID:        ids[3],
						LocationType:      locTypes[3].LocationTypeID,
						PartnerInternalID: fmt.Sprintf("partner_internal_%v", ids[3]),
						Name:              "Sub v3 1 " + ids[3],
						CreatedAt:         time.Now().AddDate(0, 0, -1), // for ordering
						UpdatedAt:         time.Now(),
					},
					{
						LocationID:        ids[4],
						LocationType:      locTypes[3].LocationTypeID,
						PartnerInternalID: fmt.Sprintf("partner_internal_%v", ids[4]),
						Name:              "Sub v3 2 " + ids[4],
						CreatedAt:         time.Now(),
						UpdatedAt:         time.Now(),
					},
				},
			},
		},
	}
	err = s.insertTreeLocation(ctx, treeLocation)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cant insert tree location: %s", err)
	}

	stepState.TreeLocation = treeLocation
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GetLocationTree(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &mpb.GetLocationTreeRequest{}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataReaderServiceClient(s.MasterMgmtConn).
		GetLocationTree(contextWithToken(s, ctx), stepState.Request.(*mpb.GetLocationTreeRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) VerifyLocationTree(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*mpb.GetLocationTreeResponse)
	actualTree := &domain.TreeLocation{}
	err := json.Unmarshal([]byte(res.Tree), actualTree)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't parse tree to json: %s\n. json: %s", err, res.Tree)
	}
	if !findRootAndCompare(stepState.TreeLocation, actualTree) {
		expected, _ := json.Marshal(stepState.TreeLocation)
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected JSON Tree is not correct:\nexpected:%s\n\n\n actual:%s", string(expected), res.Tree)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareLocationTypes(ctx context.Context) ([]*domain.LocationType, error) {
	locTypes, err := s.getLocationTypes(ctx)
	if err != nil || err == pgx.ErrNoRows {
		return nil, fmt.Errorf("can't select location types: %s", err)
	}
	if len(locTypes) < 4 {
		ctx, err = s.seedLocationTypes(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't seed location types: %s", err)
		}
		locTypes, err = s.getLocationTypes(ctx)
		if err != nil || len(locTypes) == 0 {
			return nil, fmt.Errorf("no location type: %s", err)
		}
	}
	return locTypes, nil
}

func (s *suite) getLocationTypes(ctx context.Context) ([]*domain.LocationType, error) {
	t := &repo.LocationType{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL ORDER BY LEVEL ASC LIMIT 10", strings.Join(fields, ","), t.TableName())
	rows, err := s.BobDBTrace.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locTypes []*domain.LocationType
	for rows.Next() {
		p := new(repo.LocationType)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		locTypes = append(locTypes, p.ToLocationTypeEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return locTypes, nil
}

func (s *suite) insertTreeLocation(ctx context.Context, loc *domain.TreeLocation) error {
	// Insert the current node
	command := `INSERT INTO locations (location_id, name, partner_internal_id, location_type, parent_location_id, is_archived, access_path, updated_at, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := s.BobDBTrace.Exec(ctx, command,
		loc.LocationID,
		loc.Name,
		loc.PartnerInternalID,
		loc.LocationType,
		loc.ParentLocationID,
		loc.IsArchived,
		loc.AccessPath,
		loc.UpdatedAt,
		loc.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Recursively insert the children
	for _, child := range loc.Children {
		child.AccessPath = loc.AccessPath + "/" + child.LocationID
		if child.ParentLocationID == "" {
			child.ParentLocationID = loc.LocationID
		}
		err := s.insertTreeLocation(ctx, child)
		if err != nil {
			return err
		}
	}

	return nil
}

func findRootAndCompare(expected *domain.TreeLocation, actual *domain.TreeLocation) bool {
	if expected.LocationID == actual.LocationID {
		return compareTreeLocations(expected, actual)
	}

	// Recursively search for the expected node among the children of the current node
	for _, child := range actual.Children {
		if found := findRootAndCompare(expected, child); found {
			return true
		}
	}

	return false
}

func compareTreeLocations(expected *domain.TreeLocation, actual *domain.TreeLocation) bool {
	format := "%s|%s|%s|%s|%s|%s|%s|%s|%s"
	expectedStr := fmt.Sprintf(format, expected.LocationID, expected.Name, expected.LocationType, expected.ParentLocationID, boolToStr(expected.IsArchived),
		expected.AccessPath, boolToStr(expected.IsUnauthorized), boolToStr(expected.IsLowestLevel), expected.PartnerInternalID)
	actualStr := fmt.Sprintf(format, actual.LocationID, actual.Name, actual.LocationType, actual.ParentLocationID, boolToStr(actual.IsArchived),
		actual.AccessPath, boolToStr(actual.IsUnauthorized), boolToStr(actual.IsLowestLevel), actual.PartnerInternalID)
	// Check that the current node matches
	if expectedStr != actualStr {
		fmt.Println("expected str:", expectedStr)
		fmt.Println("actual str:", actualStr)
		return false
	}

	expectedIndex := 0
	for _, actualChild := range actual.Children {
		if expectedIndex < len(expected.Children) {
			// If a matching child node is found, recursively compare the two nodes
			expectedChild := expected.Children[expectedIndex]
			if compareTreeLocations(expectedChild, actualChild) {
				expectedIndex++
			}
		}
	}

	// If all expected children have been matched, return true; otherwise, return false
	return expectedIndex == len(expected.Children)
}

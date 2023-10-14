package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	NoDisplayName       = "no-display_name-field"
	NoLevel             = "no-level-field"
	SwappedLevel        = "swapped-level"
	LevelAlreadyExisted = "level-already-existed"
)

const csvLocTypeRowFormat = "%s,%s,%d"

func (s *suite) importLocationTypeV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocationTypeV2(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportLocationTypeV2Request))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRows := s.StepState.ValidCsvRows
	expectedLocTypes := make([]string, 0, len(validRows))
	locTypeNames := make([]string, 0, len(validRows))
	for _, row := range validRows {
		r := strings.Split(row, ",")
		name := r[0]
		displayName := r[1]
		level := r[2]
		locTypeNames = append(locTypeNames, name)
		expectedLocTypes = append(expectedLocTypes, fmt.Sprintf("%s,%s,%s",
			name, displayName, level))
	}
	var (
		locTypeName        string
		locTypeDisplayName string
		level              int
	)
	query := `SELECT name, display_name, level FROM location_types WHERE name = ANY($1)
	 AND deleted_at IS NULL ORDER BY updated_at DESC LIMIT 100`
	rows, err := s.BobDBTrace.Query(ctx, query, locTypeNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	defer rows.Close()
	respLocTypes := []string{}
	for rows.Next() {
		if err := rows.Scan(&locTypeName, &locTypeDisplayName, &level); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		row := fmt.Sprintf("%s,%s,%d", locTypeName, locTypeDisplayName, level)
		respLocTypes = append(respLocTypes, row)
	}
	slices.Sort(expectedLocTypes)
	slices.Sort(respLocTypes)
	if equal := slices.Equal(expectedLocTypes, respLocTypes); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf(`location types are not updated properly
		expected: %s
		got: %s`, expectedLocTypes, respLocTypes)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkImportLocationTypeCSVErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoLevel, NoName, NoDisplayName, SwappedLevel:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				// _, _ = s.printBadRequest(ctx, stepState.ResponseErr)
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message with csv format\nexpected: %s\ngot: %s", stepState.ExpectedError, err.Error())
			}
		}
	case WrongLineValues:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message for csv line values\nexpected: %s\ngot: %s", stepState.ExpectedError, err.Error())
			}
			return s.compareBadRequest(ctx, err, s.ExpectedErrModel)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidLocationTypePayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	existingLocTypes, err := s.getExistingLocationTypes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not prepare valid location type: %v", err)
	}
	stepState.LocationTypes = existingLocTypes

	highestLevel, err := s.getHighestLocTypeLevel(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not get highest level: %v", err)
	}

	timeID := idutil.ULIDNow()
	// get highest level then +1
	r1 := fmt.Sprintf(csvLocTypeRowFormat, "org1-"+timeID, "type name-1 "+timeID, highestLevel+3)

	// get random loc type then keep the old level
	if len(stepState.LocationTypes) > 0 {
		//nolint
		rd := rand.Intn(len(stepState.LocationTypes))
		randLocType := stepState.LocationTypes[rd]
		r1 = fmt.Sprintf(csvLocTypeRowFormat, randLocType.Name, "updated "+timeID, randLocType.Level)
	}

	r2 := fmt.Sprintf(csvLocTypeRowFormat, "org2-"+timeID, "type name-2 "+timeID, highestLevel+6)
	r3 := fmt.Sprintf(csvLocTypeRowFormat, "org3-"+timeID, "type name-3 "+timeID, highestLevel+9)

	request := fmt.Sprintf(`name,display_name,level
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportLocationTypeV2Request{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInvalidLocationTypePayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	existingLocTypes, err := s.getExistingLocationTypes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not prepare valid location type: %v", err)
	}
	stepState.LocationTypes = existingLocTypes

	if len(stepState.LocationTypes) < 3 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%s", "seed more location types")
	}
	higherLocTypes := sliceutils.Filter(stepState.LocationTypes, func(locType *domain.LocationType) bool {
		return locType.Level > 1
	})
	sort.Slice(higherLocTypes, func(i, j int) bool {
		return higherLocTypes[i].Level < higherLocTypes[j].Level
	})

	timeID := idutil.ULIDNow()
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportLocationTypeV2Request{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name
				1,LocType 1`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 3, got 2"
		}
	case NoName:
		{
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`namez,display_name,level
				1,LocType 1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be name, got namez"
		}
	case NoDisplayName:
		{
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_namez,level
				1,LocType 1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be display_name, got display_namez"
		}
	case NoLevel:
		{
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,levelz
				1,LocType 1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be level, got levelz"
		}
	case LevelAlreadyExisted:
		{
			random := higherLocTypes[len(higherLocTypes)-1]

			r1 := fmt.Sprintf(csvLocTypeRowFormat, "dup level-"+timeID, "loc-type-name dup", random.Level)

			request := fmt.Sprintf(`name,display_name,level
				%s`, r1)
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(request),
			}
			// the error will show the already existed loc type in db
			stepState.ExpectedError = fmt.Sprintf("levels already existed: existing level: %d, name: %s", random.Level, random.Name)
		}
	case SwappedLevel:
		{
			highestLocType := higherLocTypes[len(higherLocTypes)-1] // x + 3 (see the seed method)
			secondLocType := higherLocTypes[len(higherLocTypes)-2]  // x -3 smaller 3 units
			// example: Swap A.6 -> A.4, B.3 -> B.5, then B will have the higher order 5 > 4
			// swap the level
			r1 := fmt.Sprintf(csvLocTypeRowFormat, secondLocType.Name, "loc-type-name 1", secondLocType.Level+2)
			r2 := fmt.Sprintf(csvLocTypeRowFormat, highestLocType.Name, "updated name in swapped", highestLocType.Level-2)

			request := fmt.Sprintf(`name,display_name,level
				%s
				%s`, r2, r1)
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(request),
			}
			// the error will show the first swapped level is randLocType1.Level
			stepState.ExpectedError = fmt.Sprintf("levels can not be swapped: existing level: %d, name: %s", highestLocType.Level, highestLocType.Name)
		}
	case WrongLineValues:
		{
			timeID := idutil.ULIDNow()
			// wrong values
			// empty display name
			r1 := fmt.Sprintf(csvLocTypeRowFormat, "org_X", "", 1)
			// wrong is_archived
			r2 := fmt.Sprintf(csvLocTypeRowFormat, "org_Z", "new name-2 "+timeID, 2)

			request := fmt.Sprintf(`name,display_name,level
				%s
				%s`, r1, r2)
			stepState.Request = &mpb.ImportLocationTypeV2Request{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "display name can not be empty",
					},
				},
			}

			stepState.InvalidCsvRows = []string{r1, r2}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getExistingLocationTypes(ctx context.Context) ([]*domain.LocationType, error) {
	locTypes := make([]*domain.LocationType, 0, 100)
	stmt :=
		`
		SELECT
			location_type_id,
			name,
			display_name,
			level
		FROM
			location_types
		where deleted_at is null and level > 0
		order by updated_at desc limit 100

		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query location type")
	}
	defer rows.Close()
	for rows.Next() {
		e := &domain.LocationType{}
		err := rows.Scan(
			&e.LocationTypeID,
			&e.Name,
			&e.DisplayName,
			&e.Level,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan location type")
		}
		locTypes = append(locTypes, e)
	}
	return locTypes, nil
}

// TODO: avoid using select max because of race condition
// TODO: check duplicating location types when seeding
func (s *suite) getHighestLocTypeLevel(ctx context.Context) (int, error) {
	query := `SELECT max(level) FROM location_types where deleted_at is null`
	max := database.Int4(0)
	err := s.BobDBTrace.QueryRow(ctx, query).Scan(&max)
	return int(max.Int), err
}

func (s *suite) seedLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertOrgLocationTypes(ctx, "org")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	rand.Seed(time.Now().UnixNano())
	//nolint
	randomLevel := rand.Intn(2147480000) // near with maximum of int4: 2,147,483,647
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not get highest level while seed: %v", err)
	}
	queryType := `INSERT INTO location_types (location_type_id, name, display_name, level, updated_at, created_at, is_archived)
	VALUES ($1, $2, $3, $4, now(), now(), false)`

	// += 3 to test the swapped value
	for i := 1; i < 14; i += 3 {
		locationTypeID := idutil.ULIDNow()
		typeName := "Seeded locType_" + locationTypeID
		_, err = s.BobDBTrace.Exec(ctx, queryType, locationTypeID, typeName, "display-"+typeName, randomLevel+i)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed location type, err: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

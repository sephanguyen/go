package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	NoLocationType            = "no-location_type-field"
	NoPartnerInternalID       = "no-partner_internal_id-field"
	NoPartnerInternalParentID = "no-partner_internal_parent_id-field"
	WrongPartnerValues        = "wrong-partner-values"
)

const locationCsvRowFormat = "%s,%s,%s,%s"

func (s *suite) importLocationV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewLocationManagementGRPCServiceClient(s.MasterMgmtConn).
		ImportLocationV2(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportLocationV2Request))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRows := s.StepState.ValidCsvRows
	expectedLocs := make([]string, 0, len(validRows))
	locNames := make([]string, 0, len(validRows))
	for _, row := range validRows {
		r := strings.Split(row, ",")
		partnerID := r[0]
		name := r[1]
		locationType := r[2]
		partnerParentID := r[3]
		locNames = append(locNames, name)
		expectedLocs = append(expectedLocs, fmt.Sprintf(locationCsvRowFormat,
			partnerID, name, locationType, partnerParentID))
	}
	locations, err := s.getExistingLocations(ctx, locNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not get locations: %s", err.Error())
	}

	actualLocs := sliceutils.Map(locations, func(l *domain.Location) string {
		return fmt.Sprintf(locationCsvRowFormat, l.PartnerInternalID, l.Name, l.LocationType, l.PartnerInternalParentID)
	})
	slices.Sort(expectedLocs)
	slices.Sort(actualLocs)
	if equal := slices.Equal(expectedLocs, actualLocs); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf(`location are not updated properly
		expected: %s
		got:      %s`, expectedLocs, actualLocs)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkImportLocationCSVErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case WrongLineValues:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message for csv line values\nexpected: %s\ngot: %s", stepState.ExpectedError, err.Error())
			}
			return s.compareBadRequest(ctx, err, s.ExpectedErrModel)
		}
	default:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message with csv format\nexpected: %s\ngot: %s", stepState.ExpectedError, err.Error())
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidLocationPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locTypes, err := s.getExistingLocationTypes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not prepare valid location: %v", err)
	}

	if len(locTypes) < 4 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("seed more location type: %v", err)
	}

	sort.Slice(locTypes, func(i, j int) bool {
		return locTypes[i].Level < locTypes[j].Level
	})
	stepState.LocationTypes = locTypes

	timeID := idutil.ULIDNow()
	rand.Seed(time.Now().Unix())
	//nolint
	randomPID, randomPID2, randomPID3, randomPID4 := fmt.Sprintf("%d", rand.Int31()), fmt.Sprintf("%d", rand.Int31()), fmt.Sprintf("%d", rand.Int31()), fmt.Sprintf("%d", rand.Int31())

	r1 := fmt.Sprintf(locationCsvRowFormat, randomPID, "location1-"+timeID, locTypes[1].Name, "")
	r2 := fmt.Sprintf(locationCsvRowFormat, randomPID2, "location2-"+timeID, locTypes[2].Name, randomPID)
	r3 := fmt.Sprintf(locationCsvRowFormat, randomPID3, "location3-"+timeID, locTypes[3].Name, randomPID2)
	r4 := fmt.Sprintf(locationCsvRowFormat, randomPID4, "location4-"+timeID, locTypes[2].Name, "")

	request := fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
	%s
	%s
	%s
	%s`, r1, r2, r3, r4)
	stepState.Request = &mpb.ImportLocationV2Request{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3, r4}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInvalidLocationPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locTypes, err := s.getExistingLocationTypes(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not prepare valid location type: %v", err)
	}

	sort.Slice(locTypes, func(i, j int) bool {
		return locTypes[i].Level < locTypes[j].Level
	})
	stepState.LocationTypes = locTypes

	if len(stepState.LocationTypes) < 3 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%s", "seed more location types")
	}

	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportLocationV2Request{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(`name,location_type
				1,brand 1`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 4, got 2"
		}
	case NoName:
		{
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_idz,name,location_type,partner_internal_parent_id
				1,brand 1,1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be partner_internal_id, got partner_internal_idz"
		}
	case NoLocationType:
		{
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_typez,partner_internal_parent_id
				1,brand 1,1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be location_type, got location_typez"
		}
	case NoPartnerInternalID:
		{
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,namez,location_type,partner_internal_parent_id
				1,brand 1,1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be name, got namez"
		}
	case NoPartnerInternalParentID:
		{
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_idz
				1,brand 1,1,1`),
			}
			stepState.ExpectedError = "csv has invalid format, column number 4 should be partner_internal_parent_id, got partner_internal_parent_idz"
		}
	case WrongLineValues:
		{
			// empty name
			r2 := fmt.Sprintf(locationCsvRowFormat, "12", "", "center", "12")
			r3 := fmt.Sprintf(locationCsvRowFormat, "org", "Location y", "12", "")
			r4 := fmt.Sprintf(locationCsvRowFormat, "location-1", "Location x", "brand", "12")
			// not a utf8 name
			r5 := fmt.Sprintf(locationCsvRowFormat, "location-1", string([]byte{0xff, 0xfe, 0xfd}), "brand", "")

			request := fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
			%s
			%s
			%s
			%s`, r2, r3, r4, r5)
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "column name is required",
					},
					{
						Field:       "Row Number: 5",
						Description: `name is not a valid UTF8 string`,
					},
				},
			}
		}
		// child location level can not be lower than parent.
	case WrongPartnerValues:
		{
			timeID := idutil.ULIDNow()
			rand.Seed(time.Now().Unix())
			//nolint
			randomPID := fmt.Sprintf("%d", rand.Int31())
			//nolint
			randomPID2 := fmt.Sprintf("%d", rand.Int31())
			//nolint
			randomPID3 := fmt.Sprintf("%d", rand.Int31())

			r1 := fmt.Sprintf(locationCsvRowFormat, randomPID, "loc1-"+timeID, locTypes[1].Name, "")
			r2 := fmt.Sprintf(locationCsvRowFormat, randomPID2, "loc2-"+timeID, locTypes[3].Name, randomPID)
			r3 := fmt.Sprintf(locationCsvRowFormat, randomPID3, "loc3-"+timeID, locTypes[2].Name, randomPID2)

			request := fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
			%s
			%s
			%s`, r1, r2, r3)
			stepState.Request = &mpb.ImportLocationV2Request{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field: "Row Number: 3",
						Description: fmt.Sprintf("%s location level (%d) must be greater than parent.\n(parent internal id: %s, location type: %s, level: %d)",
							"loc3-"+timeID, locTypes[2].Level, randomPID2, locTypes[3].Name, locTypes[3].Level),
					},
				},
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

// join to get location type name
func (s *suite) getExistingLocations(ctx context.Context, names []string) ([]*domain.Location, error) {
	loc := make([]*domain.Location, 0, 100)
	stmt :=
		`
		SELECT
			l.location_id,
			l.name,
			lt.name as location_type,
			l.partner_internal_id,
			l.partner_internal_parent_id
		FROM
			locations l
		INNER JOIN location_types lt
		ON lt.location_type_id = l.location_type
		where l.deleted_at is null and l.name = ANY($1)
		order by l.updated_at desc limit 100
		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		names,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query locations")
	}
	defer rows.Close()
	for rows.Next() {
		e := &repo.Location{}
		err := rows.Scan(
			&e.LocationID,
			&e.Name,
			&e.LocationType,
			&e.PartnerInternalID,
			&e.PartnerInternalParentID,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan location")
		}
		loc = append(loc, e.ToLocationEntity())
	}
	return loc, nil
}

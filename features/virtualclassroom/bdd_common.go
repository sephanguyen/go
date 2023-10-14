package virtualclassroom

import (
	"context"
	crypto_rand "crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

const (
	StatusNone = "none"
)

// used for test cases
func (s *suite) NewID() string {
	return idutil.ULIDNow()
}

func (s *suite) enterASchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.signedAsAccountV2(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someCenters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfLocationTypesInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLocationTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	listLocationTypes := []struct {
		locationTypeID       string
		name                 string
		parentLocationTypeID string
		archived             bool
		expected             bool
	}{
		// satisfied
		{locationTypeID: "locationtype-id-1", name: "org test", expected: true},
		{locationTypeID: "locationtype-id-2", name: "brand test", parentLocationTypeID: "locationtype-id-1", expected: true},
		{locationTypeID: "locationtype-id-3", name: "area test", parentLocationTypeID: "locationtype-id-1", expected: true},
		{locationTypeID: "locationtype-id-4", name: "center test", parentLocationTypeID: "locationtype-id-2", expected: true},
		{locationTypeID: "locationtype-id-10", name: "center-10", parentLocationTypeID: "locationtype-id-2", expected: true},

		// unsatisfied
		{locationTypeID: "locationtype-id-5", name: "test-5", archived: true},
		{locationTypeID: "locationtype-id-6", name: "test-6", parentLocationTypeID: "locationtype-id-5"},
		{locationTypeID: "locationtype-id-7", name: "test-7", parentLocationTypeID: "locationtype-id-6"},
		{locationTypeID: "locationtype-id-8", name: "test-8", parentLocationTypeID: "locationtype-id-10", archived: true},
		{locationTypeID: "locationtype-id-9", name: "test-9", parentLocationTypeID: "locationtype-id-8"},
	}

	for _, lt := range listLocationTypes {
		stmt := `INSERT INTO location_types (location_type_id,name,parent_location_type_id, is_archived,updated_at,created_at) VALUES($1,$2,$3,$4,now(),now()) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, lt.locationTypeID,
			lt.name,
			NewNullString(lt.parentLocationTypeID),
			lt.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location types with `id:%s`, %v", lt.locationTypeID, err)
		}
		if lt.expected {
			stepState.LocationTypesID = append(stepState.LocationTypesID, lt.locationTypeID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

type CreateLocation struct {
	locationID        string
	partnerInternalID string
	name              string
	parentLocationID  string
	archived          bool
	expected          bool
	accessPath        string
	locationType      string
}

func (s *suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	nBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(27))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	addedRandom := "-" + strconv.Itoa(int(nBig.Int64()))

	listLocation := []CreateLocation{
		// satisfied
		{locationID: "1" + addedRandom, partnerInternalID: "partner-internal-id-1" + addedRandom, locationType: "locationtype-id-4", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1"})},
		{locationID: "2" + addedRandom, partnerInternalID: "partner-internal-id-2" + addedRandom, locationType: "locationtype-id-5", parentLocationID: "1" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2"})},
		{locationID: "3" + addedRandom, partnerInternalID: "partner-internal-id-3" + addedRandom, locationType: "locationtype-id-6", parentLocationID: "2" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2", "3"})},
		{locationID: "7" + addedRandom, partnerInternalID: "partner-internal-id-7" + addedRandom, locationType: "locationtype-id-7", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7"})},
		// unsatisfied
		{locationID: "4" + addedRandom, partnerInternalID: "partner-internal-id-4" + addedRandom, locationType: "locationtype-id-8", parentLocationID: stepState.LocationID, archived: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4"})},
		{locationID: "5" + addedRandom, partnerInternalID: "partner-internal-id-5" + addedRandom, locationType: "locationtype-id-9", parentLocationID: "4" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "6" + addedRandom, partnerInternalID: "partner-internal-id-6" + addedRandom, locationType: "locationtype-id-1", parentLocationID: "5" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "8" + addedRandom, partnerInternalID: "partner-internal-id-8" + addedRandom, locationType: "locationtype-id-2", parentLocationID: "7" + addedRandom, archived: true, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7", "8"})},
	}

	if stepState.CreateStressTestLocation {
		listLocation = append(listLocation,
			CreateLocation{
				locationID:        "VCSTRESSTESTLOCATION",
				partnerInternalID: "partner-internal-id-99" + addedRandom,
				locationType:      "locationtype-id-4",
				parentLocationID:  stepState.LocationID,
				archived:          false,
				expected:          true,
				accessPath:        buildAccessPath(stepState.LocationID, addedRandom, []string{"VCSTRESSTESTLOCATION"}),
			},
		)
	}

	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,partner_internal_id,name,parent_location_id, is_archived, access_path, location_type) VALUES($1,$2,$3,$4,$5,$6,$7) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, l.locationID, l.partnerInternalID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived, l.accessPath,
			l.locationType)
		if err != nil {
			claims := interceptors.JWTClaimsFromContext(ctx)
			fmt.Println("claims: ", claims.Manabie.UserID, claims.Manabie.ResourcePath)
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

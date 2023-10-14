package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) locationsExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	row := s.BobDBTrace.QueryRow(ctx, `SELECT COUNT(*) FROM locations l WHERE l.deleted_at IS NULL`)
	var total int
	if err := row.Scan(&total); err != nil {
		return ctx, err
	}
	if total > 10 {
		return StepStateToContext(ctx, stepState), nil
	}

	iStmt := `INSERT INTO locations (location_id, name, location_type, partner_internal_id,
		partner_internal_parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5,  now(), now())`
	var locationTypeID string
	orgName := "org"
	queryLocationTypeOrg := fmt.Sprintf("SELECT location_type_id FROM location_types l WHERE l.name = $1 AND l.deleted_at IS NULL and l.resource_path = '%s'", fmt.Sprint(constants.ManabieSchool))
	if err := s.BobDBTrace.QueryRow(ctx, queryLocationTypeOrg, orgName).Scan(&locationTypeID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get location type org, err: %s", err)
	}
	stepState.LocationTypeOrgID = locationTypeID

	for i := 0; i < 5; i++ {
		lID := idutil.ULIDNow()
		lName := fmt.Sprintf("ロケーション %d", i)
		lPartnerID := fmt.Sprintf("location PID %d", i)
		lParentPartnerID := fmt.Sprintf("location PPID %d", i)
		_, err := s.BobDBTrace.Exec(ctx, iStmt, lID, lName, locationTypeID, lPartnerID, lParentPartnerID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location, err: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) exportLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &mpb.ExportLocationsRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataReaderServiceClient(s.MasterMgmtConn).
		ExportLocations(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsLocationsInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export locations: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.ExportLocationsResponse)
	expectedLoc, err := s.getExpectedLocationCSV(ctx)
	if err != nil {
		return ctx, fmt.Errorf("can not get expected location: %s", err)
	}
	if expectedLoc != string(resp.GetData()) {
		return ctx, fmt.Errorf("location csv is not valid:\ngot:\n%s \nexpected: \n%s", string(resp.Data), expectedLoc)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getExpectedLocationCSV(ctx context.Context) (string, error) {
	locRepo := repo.LocationRepo{}
	locs, err := locRepo.GetAllLocations(ctx, s.BobDBTrace)
	if err != nil {
		return "", err
	}
	locs = sliceutils.Filter(locs, func(l *repo.Location) bool {
		return l.LocationType.String != "org"
	})
	locStr := [][]string{{"location_id", "partner_internal_id", "name", "location_type", "partner_internal_parent_id"}}
	for _, loc := range locs {
		lineStr := []string{loc.LocationID.String, loc.PartnerInternalID.String, loc.Name.String, loc.LocationType.String,
			loc.PartnerInternalParentID.String}
		locStr = append(locStr, lineStr)
	}
	sb := strings.Builder{}
	for _, line := range locStr {
		row := sliceutils.Map(line, getEscapedStr)
		sb.WriteString(fmt.Sprintf("%s\n", strings.Join(row, ",")))
	}
	return sb.String(), nil
}

func boolToStr(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

// Get escaped tricky character like double quote
func getEscapedStr(s string) string {
	mustQuote := strings.ContainsAny(s, `"`)
	if mustQuote {
		s = strings.ReplaceAll(s, `"`, `""`)
	}

	return fmt.Sprintf(`"%s"`, s)
}

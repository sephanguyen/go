package mastermgmt

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"golang.org/x/exp/slices"
)

func (s *suite) exportLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportLocationTypesRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataReaderServiceClient(s.MasterMgmtConn).
		ExportLocationTypes(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsLocationTypesInCsv(ctx context.Context, version string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export location types: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.ExportLocationTypesResponse)
	expectedLocTypes, err := s.getExpectedLocationTypesCSV(ctx, version)
	if err != nil {
		return ctx, fmt.Errorf("can not get expected location types: %s", err)
	}
	if !compareStringSlices(expectedLocTypes, string(resp.GetData())) {
		return ctx, fmt.Errorf("location type csv is not valid:\ngot:\n%s \nexpected: \n%s", string(resp.Data), expectedLocTypes)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getExpectedLocationTypesCSV(ctx context.Context, version string) (string, error) {
	locRepo := repo.LocationTypeRepo{}
	locTypes, err := locRepo.GetAllLocationTypes(ctx, s.BobDBTrace)
	if err != nil {
		return "", err
	}
	locTypes = sliceutils.Filter(locTypes, func(l *repo.LocationType) bool {
		return l.Name.String != "org"
	})
	locTypeStr := [][]string{{"location_type_id", "name", "display_name", "parent_name"}}
	if version == "v2" {
		locTypeStr = [][]string{{"location_type_id", "name", "display_name", "level"}}
	}
	for _, loc := range locTypes {
		lineStr := []string{loc.LocationTypeID.String, loc.Name.String, loc.DisplayName.String, loc.ParentName.String}
		if version == "v2" {
			lineStr = []string{loc.LocationTypeID.String, loc.Name.String, loc.DisplayName.String, strconv.Itoa(int(loc.Level.Int))}
		}
		locTypeStr = append(locTypeStr, lineStr)
	}
	sb := strings.Builder{}
	for _, line := range locTypeStr {
		row := sliceutils.Map(line, getEscapedStr)
		sb.WriteString(fmt.Sprintf("%s\n", strings.Join(row, ",")))
	}
	return sb.String(), nil
}

func compareStringSlices(s1, s2 string) bool {
	s1Slice := strings.Split(strings.TrimSpace(s1), "\n")
	s2Slice := strings.Split(strings.TrimSpace(s2), "\n")

	slices.Sort(s1Slice)
	slices.Sort(s2Slice)
	return reflect.DeepEqual(s1Slice, s2Slice)
}

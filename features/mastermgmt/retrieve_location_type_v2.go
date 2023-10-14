package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) verifyLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*mpb.RetrieveLocationTypesV2Response)
	actualLocTypes := rsp.GetLocationTypes()
	expectedLocTypes := stepState.ValidCsvRows

	var locTypeMap = make(map[string]string, len(actualLocTypes))
	for _, lt := range actualLocTypes {
		r := fmt.Sprintf(csvLocTypeRowFormat, lt.Name, lt.DisplayName, lt.Level, "0")
		locTypeMap[lt.Name] = r
	}

	for _, l := range expectedLocTypes {
		str := strings.Split(l, ",")
		// ignore archived loc types
		if str[len(str)-1] == "0" {
			v, ok := locTypeMap[str[0]]
			if !ok || v != l {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong location types received.\nexpected: %v\ngot:%v", expectedLocTypes, locTypeMap)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveLocationTypesV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.RetrieveLocationTypesV2Request{}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataReaderServiceClient(s.MasterMgmtConn).
		RetrieveLocationTypesV2(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

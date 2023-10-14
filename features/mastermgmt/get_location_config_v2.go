package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"k8s.io/utils/strings/slices"
)

func (s *suite) initLocationConfigV2InDB(ctx context.Context, locString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	random := idutil.ULIDNow()
	// stepState.LocationConfigKeys = []string{"key-1-" + random, "key-2-" + random}
	stepState.ConfigKeys = []string{"loc-key-1-" + random, "loc-key-2-" + random}
	stepState.LocationIDs, stepState.LocationConfigKeys = parseLocationConfigString(locString)

	valueArgs := make([]interface{}, 0)
	valueStrings := make([]string, 0)
	for i, key := range stepState.ConfigKeys {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, now(), now())", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, key)
		valueArgs = append(valueArgs, "boolean")
		valueArgs = append(valueArgs, "CONFIGURATION_TYPE_LOCATION_EXTERNAL")
	}
	query := fmt.Sprintf(`
	INSERT INTO configuration_key (config_key, value_type, configuration_type, created_at, updated_at)
	 VALUES %s`, strings.Join(valueStrings, ","))
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, query, valueArgs...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations 2, err: %s", err)
	}

	orgID, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// fmt.Printf("DUMP %+v\n", orgID)

	for locationID, value := range stepState.LocationConfigKeys {
		query := `
		insert into location_configuration_value_v2 (location_config_id, config_key, location_id, config_value_type, config_value, created_at, updated_at, resource_path) 
		values 
		(uuid_generate_v4(), $1, $2, 'boolean', $3, now(), now(), $4),
		(uuid_generate_v4(), $5, $2, 'boolean', $3, now(), now(), $4)
		`
		// fmt.Printf("DUMP %+v\n", query)
		_, err = s.MasterMgmtPostgresDBTrace.Exec(ctx, query, stepState.ConfigKeys[0], locationID, value, orgID,
			stepState.ConfigKeys[1])
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations 1, err: %s", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) locationsConfigurationsAreReturnedWith(ctx context.Context, returnLocs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	_, expectReturnCfgs := parseLocationConfigString(returnLocs)

	resp := stepState.Response.(*mpb.GetConfigurationByKeysAndLocationsV2Response)
	for _, cfg := range resp.GetConfigurations() {
		if !slices.Contains(stepState.ConfigKeys, cfg.ConfigKey) {
			return ctx, fmt.Errorf("unexpected config key found %s", cfg.ConfigKey)
		}

		if expectEnable, ok := expectReturnCfgs[cfg.LocationId]; ok {
			if cfg.ConfigValue != expectEnable {
				return ctx, fmt.Errorf("config %s at location %s is expected to be %s", cfg.ConfigKey, cfg.LocationId, expectEnable)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsLocationsConfigurationsWithLocations(ctx context.Context, locs string) (context.Context, error) {
	locationIDs := strings.Split(locs, ",")

	stepState := StepStateFromContext(ctx)
	req := &mpb.GetConfigurationByKeysAndLocationsV2Request{
		Keys:        stepState.ConfigKeys,
		LocationIds: locationIDs,
	}

	stepState.Response, stepState.ResponseErr = mpb.NewExternalConfigurationServiceClient(s.Connections.MasterMgmtConn).
		GetConfigurationByKeysAndLocationsV2(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func parseLocationConfigString(loConfigStr string) ([]string, map[string]string) {
	ret := make(map[string]string, 0)
	locationConfigs := strings.Split(loConfigStr, ",")
	if len(locationConfigs) == 0 {
		return nil, nil
	}
	locationIDs := []string{}
	for _, cfg := range locationConfigs {
		cfg = strings.TrimSpace(cfg)
		cfgSegments := strings.Split(cfg, ":")

		if cfgSegments[1] == "enabled" {
			ret[cfgSegments[0]] = "true"
		} else {
			ret[cfgSegments[0]] = "false"
		}
		locationIDs = append(locationIDs, cfgSegments[0])
	}
	return locationIDs, ret
}

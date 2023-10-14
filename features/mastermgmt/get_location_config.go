package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) initLocationConfigInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	random := idutil.ULIDNow()
	stepState.ConfigKeys = []string{"key-1-" + random, "key-2-" + random}
	valueArgs := make([]interface{}, 0)
	valueStrings := make([]string, 0)
	for i, key := range stepState.ConfigKeys {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, now(), now())", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, key)
		valueArgs = append(valueArgs, "boolean")
		valueArgs = append(valueArgs, "CONFIGURATION_TYPE_EXTERNAL")
	}
	query := fmt.Sprintf(`
	INSERT INTO configuration_key (config_key, value_type, configuration_type, created_at, updated_at)
	 VALUES %s`, strings.Join(valueStrings, ","))
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, query, valueArgs...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations 2, err: %s", err)
	}

	stepState.LocationIDs = []string{idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()}
	_, err = s.MasterMgmtPostgresDBTrace.Exec(ctx, `
	WITH location_ids AS (
	select location_id from unnest(cast($1 as text[])) as location_id )
	INSERT INTO location_configuration_value (location_config_id, config_key, location_id, config_value_type, config_value, created_at, updated_at, resource_path)
	 select uuid_generate_v4() AS uuid_generate_v4, e.config_key, li.location_id, e.config_value_type, 'true', now(), now(), e.resource_path  
	 from location_ids li CROSS JOIN external_configuration_value e where  e.config_key = any($2)`, stepState.LocationIDs, stepState.ConfigKeys)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations 1, err: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLocationConfigurations(ctx context.Context, keyType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var req *mpb.GetConfigurationByKeysAndLocationsRequest
	switch keyType {
	case "empty":
		req = &mpb.GetConfigurationByKeysAndLocationsRequest{
			Keys:         []string{},
			LocationsIds: []string{},
		}
	case "existing key and locations":
		req = &mpb.GetConfigurationByKeysAndLocationsRequest{
			Keys:         stepState.ConfigKeys,
			LocationsIds: stepState.LocationIDs,
		}
	case "non existing key":
		req = &mpb.GetConfigurationByKeysAndLocationsRequest{
			Keys:         []string{"key-not-exist"},
			LocationsIds: stepState.LocationIDs,
		}
	}

	stepState.Response, stepState.ResponseErr = mpb.NewExternalConfigurationServiceClient(s.Connections.MasterMgmtConn).
		GetConfigurationByKeysAndLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnCorrectLocationConfigurations(ctx context.Context, keyType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	total := len(stepState.ConfigKeys) * len(stepState.LocationIDs)
	if keyType == "non existing key" {
		total = 0
	}

	resp := stepState.Response.(*mpb.GetConfigurationByKeysAndLocationsResponse)
	if len(resp.Configurations) != total {
		return ctx, fmt.Errorf("configurations returned not correct, need: %d", total)
	}
	return StepStateToContext(ctx, stepState), nil
}

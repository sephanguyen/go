package mastermgmt

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
)

func (s *suite) getConfigurations(ctx context.Context, keyType string, page int, limit int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var cfgKey string = ""
	switch keyType {
	case "empty":
		cfgKey = ""
	case "existing key":
		existingCfgs, _ := s.selectPaginatedConfigurations(ctx, limit, int64(page*limit))
		eKey := existingCfgs[0].ConfigKey
		cfgKey = eKey[2 : len(eKey)-3] // get substring to test search like %%
	case "non existing key":
		cfgKey = "key-not-exist_"
	}
	req := &mpb.GetConfigurationsRequest{
		Keyword: cfgKey,
		Paging: &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(limit * page),
			},
		},
	}
	stepState.Response, stepState.ResponseErr = mpb.NewConfigurationServiceClient(s.Connections.MasterMgmtConn).
		GetConfigurations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getAllConfigurations(ctx context.Context, keyType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var cfgKey string = ""
	switch keyType {
	case "empty":
		cfgKey = ""
	case "existing key":
		existingCfgs, _ := s.selectPaginatedConfigurations(ctx, 10000, 0)
		eKey := existingCfgs[0].ConfigKey
		cfgKey = eKey[2 : len(eKey)-3] // get substring to test search like %%
		fmt.Println("key: ", cfgKey)
	case "non existing key":
		cfgKey = "key-not-exist_"
	}
	req := &mpb.GetConfigurationsRequest{
		Keyword: cfgKey,
	}
	stepState.Response, stepState.ResponseErr = mpb.NewConfigurationServiceClient(s.Connections.MasterMgmtConn).
		GetConfigurations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getAllConfigurationsByService(ctx context.Context, keyType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var cfgKey string = ""
	switch keyType {
	case "empty":
		cfgKey = ""
	case "existing key":
		existingCfgs, _ := s.selectPaginatedConfigurations(ctx, 10000, 0)
		eKey := existingCfgs[0].ConfigKey
		cfgKey = eKey[2 : len(eKey)-3] // get substring to test search like %%
		fmt.Println("key: ", cfgKey)
	case "non existing key":
		cfgKey = "key-not-exist_"
	}
	req := &mpb.GetConfigurationsRequest{
		Keyword:        cfgKey,
		OrganizationId: "-2147483648",
	}
	stepState.Response, stepState.ResponseErr = mpb.NewInternalServiceClient(s.Connections.MasterMgmtConn).
		GetConfigurations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnPaginatedConfigurations(ctx context.Context, page int, limit int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not get paginated configurations %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.GetConfigurationsResponse)

	configs, err := s.selectPaginatedConfigurations(ctx, limit, int64(limit*page))
	if err != nil {
		return ctx, fmt.Errorf("can not get all configs: %s", err.Error())
	}
	remains := make([]*mpb.Configuration, 0, len(configs))
	for _, v := range resp.GetItems() {
		if !sliceutils.ContainFunc(configs, v, func(c1, c2 *mpb.Configuration) bool {
			if c1.Id == c2.Id && c1.ConfigKey == c2.ConfigKey {
				return true
			}
			return false
		}) {
			remains = append(remains, v)
		}
	}
	if len(remains) > 0 {
		return ctx, fmt.Errorf("configurations returned not correct, need: %s", remains)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someExistingConfigurationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < 10; i++ {
		query := `INSERT INTO internal_configuration_value(configuration_id, config_key, config_value, config_value_type, created_at, updated_at)
			values($1, $2, $3, $4, now(), now())`
		id := idutil.ULIDNow()
		key := "key-" + strconv.Itoa(i) + "-" + id
		value := "value-" + strconv.Itoa(i) + "-" + id
		valueType := "string"
		_, err := s.MasterMgmtDBTrace.Exec(ctx, query, id, key, value, valueType)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations, err: %s", err)
		}
	}

	for i := 0; i < 10; i++ {
		query := `INSERT INTO external_configuration_value(configuration_id, config_key, config_value, config_value_type, created_at, updated_at)
			values($1, $2, $3, $4, now(), now())`
		id := idutil.ULIDNow()
		key := "key-" + strconv.Itoa(i) + "-" + id
		value := "value-" + strconv.Itoa(i) + "-" + id
		valueType := "int"
		_, err := s.MasterMgmtDBTrace.Exec(ctx, query, id, key, value, valueType)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed external configurations, err: %s", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) configurationsExistedOnDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := s.MasterMgmtDBTrace.Exec(ctx, `INSERT INTO external_configuration_value (configuration_id, config_key, config_value, config_value_type, created_at, updated_at)
	select uuid_generate_v4(), config_key , value_type , default_value, now(), now()
	FROM configuration_key WHERE configuration_key.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL' ON CONFLICT DO NOTHING;
	INSERT INTO internal_configuration_value (configuration_id, config_key, config_value, config_value_type, created_at, updated_at)
	select uuid_generate_v4(), config_key , value_type , default_value, now(), now()
	FROM configuration_key WHERE configuration_key.configuration_type = 'CONFIGURATION_TYPE_INTERNAL' ON CONFLICT DO NOTHING;`)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed configurations, err: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnCorrectConfigurations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not get configurations %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.GetConfigurationsResponse)
	var total int
	err := s.MasterMgmtDBTrace.QueryRow(ctx, `
	SELECT count(*) from configuration_key
	where configuration_type IN ('CONFIGURATION_TYPE_INTERNAL','CONFIGURATION_TYPE_EXTERNAL')
	`).Scan(&total)
	if err != nil {
		return ctx, fmt.Errorf("can not get configuration key %s", stepState.ResponseErr.Error())
	}
	if len(resp.Items) != total {
		return ctx, fmt.Errorf("configurations returned not correct, need: %d", total)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectPaginatedConfigurations(ctx context.Context, limit int, offset int64) ([]*mpb.Configuration, error) {
	cfs := make([]*mpb.Configuration, 0, 10)
	stmt :=
		`
		SELECT 
			configuration_id,
			config_key,
			config_value,
			created_at,
			updated_at
		FROM
		internal_configuration_value
		WHERE deleted_at is null
		ORDER BY config_key DESC
		LIMIT $1 OFFSET $2
		`
	rows, err := s.MasterMgmtDBTrace.Query(
		ctx,
		stmt,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query configuration failed")
	}
	defer rows.Close()
	for rows.Next() {
		e := domain.InternalConfiguration{}
		err := rows.Scan(
			&e.ID,
			&e.ConfigKey,
			&e.ConfigValue,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan configuration failed")
		}

		cfs = append(cfs, &mpb.Configuration{Id: e.ID, ConfigKey: e.ConfigKey, ConfigValue: e.ConfigValue, CreatedAt: e.CreatedAt.String(), UpdatedAt: e.UpdatedAt.String()})
	}
	return cfs, nil
}

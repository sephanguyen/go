package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) anyOrgAndConfigkeyInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	countConfigkey := 0
	stmt := `SELECT count(*) FROM configuration_key`
	err := s.MasterMgmtPostgresDBTrace.QueryRow(ctx, stmt).Scan(&countConfigkey)
	if err != nil {
		return nil, errors.Wrap(err, "query configuration key failed")
	}

	if err != nil {
		return nil, errors.WithMessage(err, "rows.Scan count config key failed")
	}

	if countConfigkey == 0 {
		return nil, errors.Wrap(err, "no config key in DB")
	}

	countOrganization := 0
	stmt = `SELECT count(*) FROM organizations`
	err = s.MasterMgmtPostgresDBTrace.QueryRow(ctx, stmt).Scan(&countOrganization)
	if err != nil {
		return nil, errors.Wrap(err, "query organization failed")
	}

	if countOrganization == 0 {
		return nil, errors.Wrap(err, "no organization in DB")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aNewOrgInsertedIntoDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `INSERT INTO public.organizations
	(organization_id,tenant_id,"name",resource_path,domain_name,logo_url,country,created_at,updated_at,deleted_at,scrypt_signer_key,scrypt_salt_separator,scrypt_rounds,scrypt_memory_cost) 
		VALUES
	($1,$2,$3,$4,$5,NULL,NULL,now(),now(),NULL,NULL,NULL,NULL,NULL);
`
	organizationID := fmt.Sprintf("%d", time.Now().Unix())
	tenantID := "tenant-" + organizationID
	name := "org-name-" + organizationID
	resourcePath := organizationID
	domainName := "domain-" + organizationID
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, query, organizationID, tenantID, name, resourcePath, domainName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create organization, err: %s", err)
	}

	stepState.NewOrgID = organizationID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newConfigValueAddedForNewOrg(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := multierr.Combine(s.checkConfigValueByOrg(ctx, "CONFIGURATION_TYPE_INTERNAL"),
		s.checkConfigValueByOrg(ctx, "CONFIGURATION_TYPE_EXTERNAL"))

	if err != nil {
		return nil, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkConfigValueByOrg(ctx context.Context, configType string) error {
	stepState := StepStateFromContext(ctx)

	stmt := `select count(*) from configuration_key ck 
			where configuration_type = 'CONFIGURATION_TYPE_INTERNAL'
				and config_key not in (select icv.config_key from internal_configuration_value icv where icv.resource_path = $1)`
	if configType == "CONFIGURATION_TYPE_EXTERNAL" {
		stmt = `select count(*) from configuration_key ck
		where configuration_type = 'CONFIGURATION_TYPE_EXTERNAL'
			and config_key not in (select ecv.config_key from external_configuration_value ecv where ecv.resource_path = $1)`
	}
	countConfigValue := 0
	err := s.MasterMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.NewOrgID).Scan(&countConfigValue)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("check %s fail", configType))
	}

	if countConfigValue > 0 {
		return errors.Wrap(err, fmt.Sprintf("missing %s configuration", configType))
	}
	return nil
}

func (s *suite) aNewConfigKeyInsertedIntoDB(ctx context.Context, configType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `INSERT INTO configuration_key
	(value_type, default_value, configuration_type, created_at, updated_at, config_key)
	VALUES
	('string', 'off', $1, NOW(), NOW(), $2)`
	configKey := fmt.Sprintf("%d", time.Now().Unix())
	if configType == "internal" {
		configType = "CONFIGURATION_TYPE_INTERNAL"
		configKey = "internal_key_" + configKey
	}
	if configType == "external" {
		configType = "CONFIGURATION_TYPE_EXTERNAL"
		configKey = "external_key_" + configKey
	}
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, query, configType, configKey)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create config key, err: %s", err)
	}

	stepState.NewConfigKey = configKey

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newConfigValueAddedForExistingOrg(ctx context.Context, configType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.checkConfigValueByKey(ctx, configType)
	if err != nil {
		return nil, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkConfigValueByKey(ctx context.Context, configType string) error {
	stepState := StepStateFromContext(ctx)

	tableName := ""
	if configType == "internal" {
		tableName = "internal_configuration_value"
	}
	if configType == "external" {
		tableName = "external_configuration_value"
	}

	countOrg := 0
	stmt := `select count(*)
	from organizations o 
	where created_at < (select min(cv.created_at) from %s cv where cv.config_key=$1)
		and o.resource_path not in (select cv1.resource_path from %s cv1 where cv1.config_key=$2)`

	err := s.MasterMgmtPostgresDBTrace.QueryRow(ctx, fmt.Sprintf(stmt, tableName, tableName),
		stepState.NewConfigKey, stepState.NewConfigKey).Scan(&countOrg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("check %s fail", stepState.NewConfigKey))
	}

	if countOrg > 0 {
		return errors.Wrap(err, fmt.Sprintf("missing value of config key %s for any org", stepState.NewConfigKey))
	}
	return nil
}

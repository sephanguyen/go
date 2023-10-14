package mastermgmt

import (
	"context"
	"fmt"
	"time"
)

func (s *suite) createConfigurationKey(ctx context.Context, configType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// create tenant to test
	query := `INSERT INTO public.organizations
	(organization_id,tenant_id,"name",resource_path,domain_name,logo_url,country,created_at,updated_at,deleted_at,scrypt_signer_key,scrypt_salt_separator,scrypt_rounds,scrypt_memory_cost)
		VALUES
	($1,$2,$3,$4,$5,NULL,NULL,now(),now(),NULL,NULL,NULL,NULL,NULL);`

	organizationID := fmt.Sprintf("%d", time.Now().Unix())
	if configType == "internal" {
		organizationID = "audit-internal-" + organizationID
	}
	if configType == "external" {
		organizationID = "audit-external-" + organizationID
	}
	tenantID := "tenant-" + organizationID
	name := "org-name-" + organizationID
	resourcePath := organizationID
	domainName := "domain-" + organizationID
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, query, organizationID, tenantID, name, resourcePath, domainName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create organization, err: %s", err)
	}

	stepState.AuditConfigOrgID = organizationID

	// create config to test
	query = `INSERT INTO configuration_key
	(value_type, default_value, configuration_type, created_at, updated_at, config_key)
	VALUES
	('string', 'off', $1, NOW(), NOW(), $2)`
	configKey := fmt.Sprintf("audit_%d", time.Now().Unix())
	if configType == "internal" {
		configType = "CONFIGURATION_TYPE_INTERNAL"
		configKey = "internal_key_" + configKey
	}
	if configType == "external" {
		configType = "CONFIGURATION_TYPE_EXTERNAL"
		configKey = "external_key_" + configKey
	}
	_, err = s.MasterMgmtPostgresDBTrace.Exec(ctx, query, configType, configKey)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create config key, err: %s", err)
	}

	stepState.AuditConfigKey = configKey

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateValueOfConfiguration(ctx context.Context, configType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	table := ""
	if configType == "internal" {
		table = "internal_configuration_value"
	}
	if configType == "external" {
		table = "external_configuration_value"
	}

	query := `UPDATE %s
	set config_value = $1, updated_at = now()
	where config_key = $2
		and resource_path = $3`
	_, err := s.MasterMgmtPostgresDBTrace.Exec(ctx, fmt.Sprintf(query, table), "test_value", stepState.AuditConfigKey, stepState.AuditConfigOrgID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update config value, err: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkTheAuditLogRecorded(ctx context.Context, configType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	table := ""
	if configType == "internal" {
		table = "configuration_audit_internal_configuration_value"
	}
	if configType == "external" {
		table = "configuration_audit_external_configuration_value"
	}

	query := `select count(*) from %s
	where config_value = $1
		and config_key = $2
		and resource_path = $3`
	countConfigValue := 0
	err := s.MasterMgmtPostgresDBTrace.QueryRow(ctx, fmt.Sprintf(query, table), "test_value", stepState.AuditConfigKey, stepState.AuditConfigOrgID).Scan(&countConfigValue)
	if err != nil || countConfigValue == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find %s config audit record, err: %s", configType, err)
	}

	return StepStateToContext(ctx, stepState), nil
}

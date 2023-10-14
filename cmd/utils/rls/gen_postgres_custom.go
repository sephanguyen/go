package rls

import "fmt"

func buildUsingCondition(using string) string {
	if using == "" {
		return using
	}
	return fmt.Sprintf("using (%s)", using)
}
func buildWithCheckCondition(withCheck string) string {
	if withCheck == "" {
		return withCheck
	}
	return fmt.Sprintf("with check (%s)", withCheck)
}

func buildCustomPg(policy TemplatePostgresPolicy, tableName string) string {
	usingCondition := buildUsingCondition(policy.Using)
	withCheckCondition := buildWithCheckCondition(policy.WithCheck)

	return fmt.Sprintf(templateRls, policy.Name, tableName, policy.For, usingCondition, withCheckCondition)
}

func buildMultiCustomPg(postgresPolicies []TemplatePostgresPolicy, tableName string) string {
	allPolicy := ""
	for _, policy := range postgresPolicies {
		allPolicy += buildCustomPg(policy, tableName)
	}
	return dropPermissivePolicyOfMultiTenant(tableName) + allPolicy
}

func dropPermissivePolicyOfMultiTenant(table string) string {
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	return fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"
}

func (p *Postgres) genCustomPolicy(postgresPolicies *[]TemplatePostgresPolicy, tableName string) error {
	if postgresPolicies == nil {
		return nil
	}

	newMigrateFile, err := p.getNewMigrateFile(databaseName)
	if err != nil {
		return fmt.Errorf("getNewMigrateFile error %w", err)
	}

	policyContent := buildMultiCustomPg(*postgresPolicies, tableName)

	err = p.IOUtils.WriteStringFile(newMigrateFile, policyContent)

	if err != nil {
		return fmt.Errorf("file can't be write file %w", err)
	}

	return nil
}

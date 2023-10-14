------ Usermgmt Salesforce ------
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('100013', 'usermgmt-salesforce-gpkzt', 'Usermgmt Salesforce', '100013', 'usermgmt-sf', '', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

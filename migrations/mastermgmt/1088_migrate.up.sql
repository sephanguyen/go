------------ usermgmt internal -------------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('100012', 'usermgmt-internal-86vua', 'Usermgmt Internal', '100012', 'usermgmt', '', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;
------------ LMS 2.0 -------------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('-2147483622', 'lms-v2-internal-g5yjf', 'LMS 2.0', '-2147483622', 'lmsv2', '', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;
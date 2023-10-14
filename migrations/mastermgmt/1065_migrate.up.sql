------ Seiki ------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('-2147483625', 'seiki-scg-c5unw', 'scg', '-2147483625', 'scg', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant_logo/seiki-scg-logo.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

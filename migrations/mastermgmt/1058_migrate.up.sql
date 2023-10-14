------ Keishin ------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('-2147483626', 'keishin-zi2e3', 'Keishin', '-2147483626', 'manan', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant_logo/keishin-logo.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

--- Update config
UPDATE configuration_key
SET default_value = 'on'
WHERE config_key = 'user.enrollment.update_status_manual'
AND default_value = 'off';

UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'user.enrollment.update_status_manual'
AND config_value = 'off'
AND resource_path = '-2147483626';

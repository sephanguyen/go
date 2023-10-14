--- E2E HCM
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('-2147483638', 'prod-e2e-hcm-8q86x', 'E2E HCM', '-2147483638', 'e2e-hcm', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/manabie_ic_splash.png','COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483638, 'student-coach-e1e95', 'prod-e2e-hcm-8q86x') ON CONFLICT DO NOTHING;

INSERT INTO schools (school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES (-2147483638, 'E2E HCM', 'COUNTRY_JP', 1, 1, null,  false, now(), now(), false, null, null, '-2147483638') ON CONFLICT DO NOTHING;


--- Manabie demo
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('-2147483637', 'prod-manabie-demo-v2amk', 'Manabie Demo', '-2147483637', 'manabie-demo', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/manabie_ic_splash.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483637, 'student-coach-e1e95', 'prod-manabie-demo-v2amk') ON CONFLICT DO NOTHING;

INSERT INTO schools (school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES (-2147483637, 'E2E HCM', 'COUNTRY_JP', 1, 1, null,  false, now(), now(), false, null, null, '-2147483637') ON CONFLICT DO NOTHING;


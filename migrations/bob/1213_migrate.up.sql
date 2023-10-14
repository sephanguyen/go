INSERT INTO organizations ( organization_id, tenant_id,               name,       resource_path, domain_name, logo_url, country,      created_at, updated_at, deleted_at)
VALUES                    ( '-2147483639',   'prod-e2e-tokyo-2k4xb', 'E2E Tokyo', '-2147483639', 'e2e-tokyo', null,     'COUNTRY_JP', now(),      now(),      null      ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483639, 'student-coach-e1e95', 'prod-e2e-tokyo-2k4xb') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id,   name,        country,      city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES              ( -2147483639, 'E2E Tokyo', 'COUNTRY_JP', 1,       1,           null,  false,            now(),      now(),      false,    null,         null,       '-2147483639') ON CONFLICT DO NOTHING;
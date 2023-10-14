--- Create MANABIE TECH Org --- 
INSERT INTO organizations (organization_id, tenant_id,name,resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES('-2147483633', 'manabie-tech-xbglh', 'Manabie Tech', '-2147483633', 'manabie-tech', '','COUNTRY_JP', now(), now(), null ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483633, 'student-coach-e1e95', 'manabie-tech-xbglh') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES(-2147483633, 'Manabie Tech', 'COUNTRY_JP', 1,1, null,false,now(), now(), false, null, null,'-2147483633') ON CONFLICT DO NOTHING;

--- Create Manabie Empty Tenant Org --- 
INSERT INTO organizations (organization_id, tenant_id,name,resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES('-2147483632', 'manabie-kael-6ys95', 'Manabie Data Leak check', '-2147483632', 'manabie-kael', '','COUNTRY_JP', now(), now(), null ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483632, 'student-coach-e1e95', 'manabie-kael-6ys95') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES(-2147483632, 'Manabie Data Leak check', 'COUNTRY_JP', 1,1, null,false,now(), now(), false, null, null,'-2147483632') ON CONFLICT DO NOTHING;
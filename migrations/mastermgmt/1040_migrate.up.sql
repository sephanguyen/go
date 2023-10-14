INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
select '100000', 'e2e-architecture-29vl6', 'E2E Architecture', '100000', 'e2e-architecture', '', 'COUNTRY_JP', now(), now(), null
where not exists (select 1 from organizations o1 where o1.resource_path = '100000');

ALTER TABLE public.configuration_key ALTER COLUMN configuration_type SET NOT NULL;

ALTER TABLE public.configuration_key ALTER COLUMN default_value SET DEFAULT '';
ALTER TABLE public.configuration_key ALTER COLUMN default_value SET NOT NULL;

ALTER TABLE public.internal_configuration_value ALTER COLUMN config_value SET DEFAULT '';
ALTER TABLE public.internal_configuration_value ALTER COLUMN config_value SET NOT NULL;

ALTER TABLE public.external_configuration_value ALTER COLUMN config_value SET DEFAULT '';
ALTER TABLE public.external_configuration_value ALTER COLUMN config_value SET NOT NULL;

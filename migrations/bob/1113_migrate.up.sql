ALTER TABLE public.organizations
    DROP CONSTRAINT IF EXISTS organizations__pk;
ALTER TABLE public.organizations
    ADD CONSTRAINT organizations__pk PRIMARY KEY (organization_id);
ALTER TABLE public.organizations
    DROP CONSTRAINT IF EXISTS organizations__tenant_id__un;
ALTER TABLE public.organizations
    ADD CONSTRAINT organizations__tenant_id__un UNIQUE (tenant_id);
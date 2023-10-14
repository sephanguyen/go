CREATE TABLE IF NOT EXISTS public.organization_auths
(
    organization_id integer,
    auth_project_id text,
    auth_tenant_id  text
);

ALTER TABLE public.organization_auths
    DROP CONSTRAINT IF EXISTS organization_auths__pk;
ALTER TABLE public.organization_auths
    ADD CONSTRAINT organization_auths__pk PRIMARY KEY (organization_id, auth_project_id, auth_tenant_id);

CREATE INDEX IF NOT EXISTS organization_auths__organization_id__idx
    ON public.organization_auths (CAST(organization_id AS text));
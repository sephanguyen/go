CREATE TABLE IF NOT EXISTS public.permission (
    permission_id TEXT NOT NULL,
    permission_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);
CREATE POLICY rls_permission ON public.permission
    USING (permission_check(resource_path, 'permission'))
    WITH CHECK (permission_check(resource_path, 'permission'));
CREATE POLICY rls_permission_restrictive ON "permission"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'permission'))
    WITH CHECK (permission_check(resource_path, 'permission'));
ALTER TABLE public.permission ENABLE ROW LEVEL security;
ALTER TABLE public.permission FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.permission_role (
    permission_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id, resource_path)
);
CREATE POLICY rls_permission_role ON public.permission_role
    USING (permission_check(resource_path, 'permission_role'))
    WITH CHECK (permission_check(resource_path, 'permission_role'));
CREATE POLICY rls_permission_role_restrictive ON "permission_role"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'permission_role'))
    WITH CHECK (permission_check(resource_path, 'permission_role'));
ALTER TABLE public.permission_role ENABLE ROW LEVEL security;
ALTER TABLE public.permission_role FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.granted_role_access_path (
    granted_role_id TEXT NOT NULL,
    location_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id)
);
CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path
    USING (permission_check(resource_path, 'granted_role_access_path'))
    WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
CREATE POLICY rls_granted_role_access_path_restrictive ON "granted_role_access_path" AS RESTRICTIVE TO PUBLIC
    USING (permission_check(resource_path,'granted_role_access_path'))
    WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role_access_path FORCE ROW LEVEL security;

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


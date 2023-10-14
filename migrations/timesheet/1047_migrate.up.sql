CREATE TABLE IF NOT EXISTS public.permission_role (
    permission_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,
    CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id, resource_path)
);
CREATE POLICY rls_permission_role ON public.permission_role USING (permission_check(resource_path, 'permission_role')) WITH CHECK (permission_check(resource_path, 'permission_role'));
ALTER TABLE public.permission_role ENABLE ROW LEVEL security;
ALTER TABLE public.permission_role FORCE ROW LEVEL security;

CREATE POLICY rls_permission_role_restrictive ON "permission_role" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'permission_role'))
WITH CHECK (permission_check(resource_path, 'permission_role'));
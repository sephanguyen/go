CREATE TABLE IF NOT EXISTS public.granted_role_access_path (
    granted_role_id TEXT NOT NULL,
    location_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id,location_id)
);
CREATE POLICY rls_granted_role_access_path ON "granted_role_access_path" USING (permission_check(resource_path, 'granted_role_access_path')) WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
CREATE POLICY rls_granted_role_access_path_restrictive ON "granted_role_access_path" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'granted_role_access_path')) with check (permission_check(resource_path, 'granted_role_access_path'));

ALTER TABLE "granted_role_access_path" ENABLE ROW LEVEL security;
ALTER TABLE "granted_role_access_path" FORCE ROW LEVEL security;

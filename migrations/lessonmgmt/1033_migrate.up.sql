CREATE TABLE IF NOT EXISTS public.granted_role (
    granted_role_id TEXT NOT NULL UNIQUE,
    user_group_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),
    
    CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id)
);
CREATE POLICY rls_granted_role ON "granted_role" USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
CREATE POLICY rls_granted_role_restrictive ON "granted_role" AS RESTRICTIVE FOR ALL TO PUBLIC USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));

ALTER TABLE "granted_role" ENABLE ROW LEVEL security;
ALTER TABLE "granted_role" FORCE ROW LEVEL security;

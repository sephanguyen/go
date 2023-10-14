CREATE TABLE IF NOT EXISTS public.role (
    role_id TEXT NOT NULL,
    role_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__role PRIMARY KEY (role_id,resource_path)
);
CREATE POLICY rls_role ON "role" USING (permission_check(resource_path, 'role')) WITH CHECK (permission_check(resource_path, 'role'));
CREATE POLICY rls_role_restrictive ON "role" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'role')) with check (permission_check(resource_path, 'role'));

ALTER TABLE "role" ENABLE ROW LEVEL security;
ALTER TABLE "role" FORCE ROW LEVEL security;

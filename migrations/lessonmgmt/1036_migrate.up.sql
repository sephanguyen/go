CREATE TABLE public.permission_role (
	permission_id text NOT NULL,
	role_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT permission_role__pk PRIMARY KEY (permission_id,role_id,resource_path)
);

CREATE POLICY rls_permission_role ON "permission_role" USING (permission_check(resource_path, 'permission_role')) WITH CHECK (permission_check(resource_path, 'permission_role'));
CREATE POLICY rls_permission_role_restrictive ON "permission_role" AS RESTRICTIVE FOR ALL TO PUBLIC USING (permission_check(resource_path, 'permission_role')) WITH CHECK (permission_check(resource_path, 'permission_role'));

ALTER TABLE "permission_role" ENABLE ROW LEVEL security;
ALTER TABLE "permission_role" FORCE ROW LEVEL security;

CREATE TABLE public.user_group (
	user_group_id text NOT NULL,
	user_group_name text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	org_location_id text NULL,
	is_system bool NULL DEFAULT false,
	CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);

CREATE POLICY rls_user_group ON "user_group" USING (permission_check(resource_path, 'user_group')) WITH CHECK (permission_check(resource_path, 'user_group'));
CREATE POLICY rls_user_group_restrictive ON "user_group" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'user_group')) with check (permission_check(resource_path, 'user_group'));

ALTER TABLE "user_group" ENABLE ROW LEVEL security;
ALTER TABLE "user_group" FORCE ROW LEVEL security;

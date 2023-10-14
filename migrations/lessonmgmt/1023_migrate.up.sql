-- public.user_basic_info definition
CREATE TABLE public.user_basic_info (
	user_id text NOT NULL,
	"name" text NULL,
	first_name text NULL,
	last_name text NULL,
	full_name_phonetic text NULL,
	first_name_phonetic text NULL,
	last_name_phonetic text NULL,
	current_grade int2 NULL,
	grade_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	email text NULL,
	CONSTRAINT pk__user_basic_info PRIMARY KEY (user_id)
);

CREATE POLICY rls_user_basic_info ON "user_basic_info" USING (permission_check(resource_path, 'user_basic_info')) WITH CHECK (permission_check(resource_path, 'user_basic_info'));
CREATE POLICY rls_user_basic_info_restrictive ON "user_basic_info" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'user_basic_info')) with check (permission_check(resource_path, 'user_basic_info'));

ALTER TABLE "user_basic_info" ENABLE ROW LEVEL security;
ALTER TABLE "user_basic_info" FORCE ROW LEVEL security;

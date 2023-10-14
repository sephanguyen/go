CREATE TABLE IF NOT EXISTS public.ac_test_template_11_4 (
	ac_test_template_11_4_id text NOT NULL,
	"name" text NULL,
	"owners" text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__ac_test_template_11_4 PRIMARY KEY (ac_test_template_11_4_id)
);
CREATE POLICY rls_ac_test_template_11_4 ON public.ac_test_template_11_4 USING (permission_check(resource_path, 'ac_test_template_11_4')) WITH CHECK (permission_check(resource_path, 'ac_test_template_11_4'));
CREATE POLICY rls_ac_test_template_11_4_restrictive ON "ac_test_template_11_4" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_test_template_11_4')) with check (permission_check(resource_path, 'ac_test_template_11_4'));
ALTER TABLE public.ac_test_template_11_4 ENABLE ROW LEVEL security;
ALTER TABLE public.ac_test_template_11_4 FORCE ROW LEVEL security;


CREATE TABLE public.ac_test_template_11_4_access_paths (
	ac_test_template_11_4_id text NOT NULL,
	location_id text NOT NULL,
	access_path text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT ac_test_template_11_4_access_paths_pk PRIMARY KEY (ac_test_template_11_4_id, location_id)
);
CREATE POLICY rls_ac_test_template_11_4_access_paths ON public.ac_test_template_11_4_access_paths USING (permission_check(resource_path, 'ac_test_template_11_4_access_paths')) WITH CHECK (permission_check(resource_path, 'ac_test_template_11_4_access_paths'));
CREATE POLICY rls_ac_test_template_11_4_access_paths_restrictive ON "ac_test_template_11_4_access_paths" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_test_template_11_4_access_paths')) with check (permission_check(resource_path, 'ac_test_template_11_4_access_paths'));
ALTER TABLE public.ac_test_template_11_4_access_paths ENABLE ROW LEVEL security;
ALTER TABLE public.ac_test_template_11_4_access_paths FORCE ROW LEVEL security;


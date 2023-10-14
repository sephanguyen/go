CREATE TABLE IF NOT EXISTS public.ac_test_template_1 (
	ac_test_template_1_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__ac_test_template_1 PRIMARY KEY (ac_test_template_1_id)
);
CREATE POLICY rls_ac_test_template_1 ON public.ac_test_template_1 USING (permission_check(resource_path, 'ac_test_template_1')) WITH CHECK (permission_check(resource_path, 'ac_test_template_1'));
CREATE POLICY rls_ac_test_template_1_restrictive ON "ac_test_template_1" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_test_template_1')) with check (permission_check(resource_path, 'ac_test_template_1'));
ALTER TABLE public.ac_test_template_1 ENABLE ROW LEVEL security;
ALTER TABLE public.ac_test_template_1 FORCE ROW LEVEL security;



CREATE TABLE public.ac_test_template_1_access_paths (
	ac_test_template_1_id text NOT NULL,
	location_id text NOT NULL,
	access_path text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT ac_test_template_1_access_paths_pk PRIMARY KEY (ac_test_template_1_id, location_id)
);
CREATE POLICY rls_ac_test_template_1_access_paths ON public.ac_test_template_1_access_paths USING (permission_check(resource_path, 'ac_test_template_1_access_paths')) WITH CHECK (permission_check(resource_path, 'ac_test_template_1_access_paths'));
CREATE POLICY rls_ac_test_template_1_access_paths_restrictive ON "ac_test_template_1_access_paths" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_test_template_1_access_paths')) with check (permission_check(resource_path, 'ac_test_template_1_access_paths'));
ALTER TABLE public.ac_test_template_1_access_paths ENABLE ROW LEVEL security;
ALTER TABLE public.ac_test_template_1_access_paths FORCE ROW LEVEL security;


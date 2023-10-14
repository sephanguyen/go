CREATE TABLE IF NOT EXISTS public.ac_test_template_4 (
	ac_test_template_4_id text NOT NULL,
	"name" text NULL,
	"owners" text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__ac_test_template_4 PRIMARY KEY (ac_test_template_4_id)
);
CREATE POLICY rls_ac_test_template_4 ON public.ac_test_template_4 USING (permission_check(resource_path, 'ac_test_template_4')) WITH CHECK (permission_check(resource_path, 'ac_test_template_4'));
CREATE POLICY rls_ac_test_template_4_restrictive ON "ac_test_template_4" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_test_template_4')) with check (permission_check(resource_path, 'ac_test_template_4'));
ALTER TABLE public.ac_test_template_4 ENABLE ROW LEVEL security;
ALTER TABLE public.ac_test_template_4 FORCE ROW LEVEL security;
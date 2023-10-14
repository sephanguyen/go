CREATE TABLE IF NOT EXISTS public.ac_hasura_test_template_1 (
	ac_hasura_test_template_1_id text NOT NULL,
	"name" text NULL,
	"location_id" text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT NOW(),
	updated_at timestamptz NOT NULL DEFAULT NOW(),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__ac_hasura_test_template_1 PRIMARY KEY (ac_hasura_test_template_1_id)
);
CREATE POLICY rls_ac_hasura_test_template_1 ON public.ac_hasura_test_template_1 USING (permission_check(resource_path, 'ac_hasura_test_template_1')) WITH CHECK (permission_check(resource_path, 'ac_hasura_test_template_1'));
CREATE POLICY rls_ac_hasura_test_template_1_restrictive ON "ac_hasura_test_template_1" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,'ac_hasura_test_template_1')) with check (permission_check(resource_path, 'ac_hasura_test_template_1'));
ALTER TABLE public.ac_hasura_test_template_1 ENABLE ROW LEVEL security;
ALTER TABLE public.ac_hasura_test_template_1 FORCE ROW LEVEL security;


DO
$do$
BEGIN
	IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='hasura') THEN
		GRANT DELETE ON ac_hasura_test_template_1 to hasura;
	END IF;
END
$do$

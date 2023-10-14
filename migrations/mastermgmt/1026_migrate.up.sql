ALTER TABLE public.ac_test_template_4 ALTER COLUMN created_at SET DEFAULT NOW();
ALTER TABLE public.ac_test_template_4 ALTER COLUMN updated_at SET DEFAULT NOW();

DO
$do$
BEGIN
	IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='hasura') THEN
		GRANT DELETE ON ac_test_template_4 to hasura;
	END IF;
END
$do$
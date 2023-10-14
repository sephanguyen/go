--for internal
CREATE TABLE IF NOT exists public.configuration_audit_internal_configuration_value (
	configuration_id text NOT NULL,
	config_key text NOT NULL,
	config_value text NOT NULL DEFAULT ''::text,
	config_value_type text NOT NULL DEFAULT 'string'::text,
	last_editor text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc' :: text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc' :: text, now()),
	deleted_at timestamptz NULL,
	resource_path text NULL,
	action_time timestamptz NOT NULL DEFAULT timezone('utc' :: text, now())
);

CREATE OR REPLACE FUNCTION audit_internal_configuration_value_fn() RETURNS TRIGGER
AS $$
    BEGIN
        INSERT INTO configuration_audit_internal_configuration_value select new.*, now();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_internal_configuration_value on public.internal_configuration_value;
CREATE TRIGGER audit_internal_configuration_value AFTER UPDATE ON public.internal_configuration_value FOR EACH ROW EXECUTE PROCEDURE audit_internal_configuration_value_fn();

--for external
CREATE TABLE IF NOT exists public.configuration_audit_external_configuration_value (
	configuration_id text NOT NULL,
	config_key text NOT NULL,
	config_value text NOT NULL DEFAULT ''::text,
	config_value_type text NOT NULL DEFAULT 'string'::text,
	last_editor text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc' :: text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc' :: text, now()),
	deleted_at timestamptz NULL,
	resource_path text NULL,
	action_time timestamptz NOT NULL DEFAULT timezone('utc' :: text, now())
);

CREATE OR REPLACE FUNCTION audit_external_configuration_value_fn() RETURNS TRIGGER
AS $$
    BEGIN
        INSERT INTO configuration_audit_external_configuration_value select new.*, now();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_external_configuration_value on public.external_configuration_value;
CREATE TRIGGER audit_external_configuration_value AFTER UPDATE ON public.external_configuration_value FOR EACH ROW EXECUTE PROCEDURE audit_external_configuration_value_fn();

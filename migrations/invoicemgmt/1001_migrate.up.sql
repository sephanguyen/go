CREATE SCHEMA IF NOT EXISTS public;

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_with_oids = false;

CREATE OR REPLACE FUNCTION autoFillResourcePath() RETURNS TEXT 
AS $$
DECLARE
	resource_path text;
BEGIN
	resource_path := current_setting('permission.resource_path', 't');

	RETURN resource_path;
END $$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS public.invoice (
    invoice_id integer NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    student_id text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_pk PRIMARY KEY (invoice_id)
);

CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;

CREATE POLICY rls_invoice ON "invoice" using (permission_check(resource_path, 'invoice')) with check (permission_check(resource_path, 'invoice'));

ALTER TABLE "invoice" ENABLE ROW LEVEL security;
ALTER TABLE "invoice" FORCE ROW LEVEL security;

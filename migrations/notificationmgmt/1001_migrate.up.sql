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


CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;
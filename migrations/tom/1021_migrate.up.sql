CREATE OR REPLACE FUNCTION autoFillResourcePath() RETURNS TEXT 
AS $$
DECLARE
		resource_path text;
BEGIN
	resource_path := current_setting('permission.resource_path', 't');

	RETURN resource_path;
END $$ LANGUAGE plpgsql;

do $$
declare selectrow record;
begin
for selectrow in
    select 'ALTER TABLE ONLY '|| T.mytable || ' ALTER COLUMN resource_path SET DEFAULT autoFillResourcePath();' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public') t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;
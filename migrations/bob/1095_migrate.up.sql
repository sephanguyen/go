do $$
declare selectrow record;
begin
for selectrow in
    select 'alter table '|| T.mytable || ' add column if not exists resource_path text;' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public') t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;


CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;


do $$
declare selectrow record;
begin
for selectrow in
    select 'CREATE POLICY rls_' || T.mytable || ' ON ' || T.mytable || ' using (permission_check(resource_path,' || T.mytable || '::text)) with check (permission_check(resource_path,'|| T.mytable || '::text));' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public') t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;

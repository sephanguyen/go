do $$
declare selectrow record;
begin
for selectrow in
    select 'alter table '|| T.mytable || ' add column if not exists deleted_at timestamp with time zone;' as script
   from (select tablename as mytable from  pg_tables where schemaname  ='public') t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;

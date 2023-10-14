do $$
declare selectrow record;
exclude_table text[]= '{
	debezium_signal,
	dbz_signals,
	organizations,
	organization_auths,
	schools,
	cities,
	districts,
	configs,
	conversion_tasks,
	districts,
	student_event_logs,
	schema_migrations}';
begin
for selectrow in
     select 'DROP POLICY IF EXISTS rls_' || T.mytable || '_restrictive ON "' || T.mytable || '" ' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public' AND NOT (tablename = ANY(exclude_table)) ) t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;

do $$
declare selectrow record;
exclude_table text[]= '{
	debezium_signal,
	dbz_signals,
	organizations,
	organization_auths,
	schools,
	cities,
	districts,
	configs,
	conversion_tasks,
	districts,
	student_event_logs,
	schema_migrations}';
begin
for selectrow in
    select 'CREATE POLICY rls_' || t.mytable || '_restrictive ON "' || T.mytable || '" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path,''' || t.mytable || ''')) with check (permission_check(resource_path, '''|| T.mytable || '''));' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public' AND NOT (tablename = ANY(exclude_table))) t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;
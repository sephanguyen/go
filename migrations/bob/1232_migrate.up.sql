update users set resource_path='UNKNOWN' where resource_path is null;
update users_groups set resource_path='UNKNOWN' where resource_path is null;
update groups set resource_path='UNKNOWN' where resource_path is null;

do $$
declare 
	selectrow record;
	exclude_table text[]= '{
  debezium_signal,
  dbz_signals,
  organization_auths,
  hubs,
  promotions,
  students_assigned_coaches,
  packages,
  student_submission_scores,
  students_study_plans_weekly,
  student_statistics,
  student_orders,
  student_comments,
  student_submissions,
  student_assignments,
  cities,
  cod_orders,
  questions,
  questions_tagged_learning_objectives,
  package_items,
  notifications,
  student_subscriptions,
  students_bk,
  quizsets,
  promotion_rules,
  coaches,
  schools,
  notification_messages,
  notification_targets,
  ios_transactions,
  plans,
  conversion_tasks,
  student_questions,
  school_configs,
  configs,
  user_notifications,
  districts,
  apple_users,
  students_topics_overdue,
  tutors,
  assignments,
  hub_tours,
  billing_histories,
  schema_migrations
}';
begin
for selectrow in
    select 'alter table '|| T.mytable || ' alter column resource_path set not null;' as script
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
    select 'alter table "'|| T.mytable || '" add column if not exists resource_path text;' as script
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
    select 'ALTER TABLE ONLY "'|| T.mytable || '" ALTER COLUMN resource_path SET DEFAULT autoFillResourcePath();' as script
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
     select 'DROP POLICY IF EXISTS rls_' || T.mytable || ' ON "' || T.mytable || '" ' as script
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
    select 'CREATE POLICY rls_' || t.mytable || ' ON "' || T.mytable || '" using (permission_check(resource_path,''' || t.mytable || ''')) with check (permission_check(resource_path, '''|| T.mytable || '''));' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public' AND NOT (tablename = ANY(exclude_table))) t
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
	select 'alter table "'|| T.mytable || '" FORCE ROW LEVEL security;' as script
	from (select tablename as mytable from pg_tables where schemaname = 'public' AND NOT (tablename = ANY(exclude_table))) t
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
	select 'alter table "'|| T.mytable || '" ENABLE ROW LEVEL security;' as script
	from (select tablename as mytable from pg_tables where schemaname = 'public' AND NOT (tablename = ANY(exclude_table))) t
    loop
begin
execute selectrow.script;
end;
end loop;
end;
$$;

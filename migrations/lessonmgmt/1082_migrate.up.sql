CREATE TABLE IF NOT EXISTS dbz_signals (
    id TEXT PRIMARY KEY, 
    type TEXT, 
    data TEXT
);


CREATE TABLE IF NOT EXISTS public.debezium_heartbeat (
  id INTEGER PRIMARY KEY,
  updated_at TIMESTAMPTZ
);


do
$do$
begin
   if not exists (select from pg_publication where pubname='debezium_publication') then
      create publication debezium_publication;
   end if;
end
$do$;


ALTER PUBLICATION debezium_publication SET TABLE 
public.dbz_signals,
public.debezium_heartbeat,
public.lessons,
public.lessons_teachers,
public.lessons_courses;

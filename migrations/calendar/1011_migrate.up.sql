
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
public.day_info,
public.scheduler;
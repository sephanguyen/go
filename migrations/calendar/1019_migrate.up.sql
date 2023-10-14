CREATE OR REPLACE FUNCTION public.permission_check(resource_path text, table_name text)
 RETURNS boolean
 LANGUAGE sql
 STABLE
AS $function$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$function$;

DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='alloydb_publication') THEN
      CREATE PUBLICATION alloydb_publication;
   END IF;
END
$do$;

DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'scheduler') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.scheduler;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'day_info') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.day_info;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'day_type') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.day_type;
  END IF;
END;
$$;

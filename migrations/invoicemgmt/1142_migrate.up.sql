CREATE OR REPLACE FUNCTION public.permission_check(resource_path text, table_name text)
 RETURNS boolean
 LANGUAGE sql
 STABLE
AS $function$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$function$;

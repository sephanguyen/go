DROP POLICY IF EXISTS rls_locations_location on locations; 

CREATE POLICY rls_locations ON public.locations
    AS permissive
    for all
    using (permission_check(resource_path, 'locations'::text))
with check (permission_check(resource_path, 'locations'::text));
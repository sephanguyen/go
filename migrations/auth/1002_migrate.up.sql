DROP POLICY rls_users_update_location ON public.users;
DROP POLICY rls_users_select_location ON public.users;
DROP POLICY rls_users_permission_v4 ON public.users;
DROP POLICY rls_users_delete_location ON public.users;
DROP POLICY rls_users_insert_location ON public.users;
CREATE POLICY rls_users ON public.users USING (public.permission_check(resource_path, 'users'::text)) WITH CHECK (public.permission_check(resource_path, 'users'::text));

DROP POLICY rls_user_access_paths_location ON public.user_access_paths;
CREATE POLICY rls_user_access_paths ON public.user_access_paths USING (public.permission_check(resource_path, 'user_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'user_access_paths'::text));

DROP POLICY rls_locations_location ON public.locations;
CREATE POLICY rls_locations ON public.locations USING (public.permission_check(resource_path, 'locations'::text)) WITH CHECK (public.permission_check(resource_path, 'locations'::text));


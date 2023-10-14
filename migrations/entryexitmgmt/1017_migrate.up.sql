ALTER TABLE public.granted_role_access_path DROP CONSTRAINT IF EXISTS fk__granted_role_access_path__granted_role_id;
ALTER TABLE public.granted_role_access_path DROP CONSTRAINT IF EXISTS fk__granted_role_access_path__location_id;

ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_users_fk;
ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_locations_fk;

ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_id;
ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_group_id;

ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__user_group_id;
ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__role_id;

ALTER TABLE public.permission_role DROP CONSTRAINT IF EXISTS fk__permission_role__permission_id;
ALTER TABLE public.permission_role DROP CONSTRAINT IF EXISTS fk__permission_role__role_id;
ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_locations_fk;
ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_users_fk;

ALTER TABLE public.user_group DROP CONSTRAINT IF EXISTS fk__user_group__org_location_id;

ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_group_id;
ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_id;

ALTER TABLE public.users_groups DROP CONSTRAINT IF EXISTS fk__users_groups__group_id;
ALTER TABLE public.users_groups DROP CONSTRAINT IF EXISTS fk__users_groups__user_id;

ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS granted_role__role_id__resource_path__fk;
ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__user_group_id;

ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_users_fk;
ALTER TABLE public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_locations_fk;

ALTER TABLE public.user_group DROP CONSTRAINT IF EXISTS fk__user_group__org_location_id;
ALTER TABLE IF EXISTS public.users DROP CONSTRAINT IF EXISTS users__email_key;
ALTER TABLE IF EXISTS public.users DROP CONSTRAINT IF EXISTS users__facebook_id__key;


ALTER TABLE IF EXISTS public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_locations_fk;
ALTER TABLE IF EXISTS public.user_access_paths DROP CONSTRAINT IF EXISTS user_access_paths_users_fk;
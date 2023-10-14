-- user table
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS user_role TEXT DEFAULT NULL;

DROP INDEX IF EXISTS users__user_role__idx;
CREATE INDEX users__user_role__idx ON public.users (user_role);

-- user_basic_info
ALTER TABLE public.user_basic_info ADD COLUMN IF NOT EXISTS user_role TEXT DEFAULT NULL;

DROP INDEX IF EXISTS user_basic_info__user_role__idx;
CREATE INDEX user_basic_info__user_role__idx ON public.user_basic_info (user_role);

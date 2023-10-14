ALTER TABLE public.info_notifications_access_paths ADD COLUMN IF NOT EXISTS created_user_id TEXT default NULL;

ALTER TABLE public.users ADD COLUMN IF NOT EXISTS last_login_date timestamp with time zone;
UPDATE public.users SET last_login_date = current_timestamp WHERE device_token IS NOT NULL AND user_group = 'USER_GROUP_STUDENT' AND last_login_date IS NULL;
UPDATE public.users u SET last_login_date = current_timestamp FROM public.student_event_logs s WHERE u.user_id = s.student_id AND u.user_group = 'USER_GROUP_STUDENT' AND u.last_login_date IS NULL;

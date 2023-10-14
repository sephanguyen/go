CREATE INDEX IF NOT EXISTS users_info_notification_user_id_idx ON public.users_info_notifications(user_id);

CREATE INDEX IF NOT EXISTS users_info_notification_notification_id_idx ON public.users_info_notifications(notification_id);

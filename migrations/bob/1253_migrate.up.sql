ALTER TABLE public.notification_locations ALTER COLUMN deleted_at DROP NOT NULL;
ALTER TABLE public.notification_locations RENAME TO info_notifications_access_paths;

ALTER POLICY rls_notification_locations ON public.info_notifications_access_paths RENAME TO rls_info_notifications_access_paths;

ALTER POLICY rls_info_notifications_access_paths ON "info_notifications_access_paths"
USING (permission_check(resource_path, 'info_notifications_access_paths'))
WITH CHECK (permission_check(resource_path, 'info_notifications_access_paths'));
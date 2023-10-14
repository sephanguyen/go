CREATE TABLE IF NOT EXISTS notification_class_members(
  student_id TEXT NOT NULL,
  class_id TEXT NOT NULL,
  start_at timestamp with time zone,
  end_at timestamp with time zone,
  created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
  updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
  resource_path text COLLATE pg_catalog."default" DEFAULT autofillresourcepath()
);

CREATE POLICY rls_notification_class_members ON "notification_class_members"
USING (permission_check(resource_path, 'notification_class_members'))
WITH CHECK (permission_check(resource_path, 'notification_class_members'));

ALTER TABLE "notification_class_members" ENABLE ROW LEVEL security;
ALTER TABLE "notification_class_members" FORCE ROW LEVEL security;

ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS excluded_generic_receiver_ids TEXT[];
ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS generic_receiver_ids TEXT[];
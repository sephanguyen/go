ALTER TABLE public.notification_class_members ADD COLUMN IF NOT EXISTS course_id TEXT NOT NULL;
ALTER TABLE public.notification_class_members ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE public.notification_class_members DROP CONSTRAINT IF EXISTS pk__notification_class_members;
ALTER TABLE public.notification_class_members ADD CONSTRAINT pk__notification_class_members PRIMARY KEY (student_id, course_id, location_id);
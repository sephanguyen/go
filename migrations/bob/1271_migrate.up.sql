-- notification_student_courses
ALTER TABLE public.notification_student_courses ADD COLUMN IF NOT EXISTS location_id TEXT DEFAULT NULL;
ALTER TABLE public.notification_student_courses DROP CONSTRAINT IF EXISTS pk__notification_student_courses;
ALTER TABLE public.notification_student_courses ALTER COLUMN subscription_id DROP NOT NULL;
ALTER TABLE public.notification_student_courses ADD CONSTRAINT pk__notification_student_courses PRIMARY KEY (course_id, student_id);

-- notification_class_members
ALTER TABLE public.notification_class_members ADD COLUMN IF NOT EXISTS location_id TEXT DEFAULT NULL;
ALTER TABLE public.notification_class_members ADD CONSTRAINT pk__notification_class_members PRIMARY KEY (class_id, student_id);
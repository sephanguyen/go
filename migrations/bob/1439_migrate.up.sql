ALTER TABLE notification_student_courses ADD COLUMN student_course_id TEXT;

ALTER TABLE public.notification_class_members DROP CONSTRAINT IF EXISTS pk__notification_class_members;
ALTER TABLE public.notification_class_members ADD CONSTRAINT pk__notification_class_members PRIMARY KEY (student_id, course_id, class_id, location_id);

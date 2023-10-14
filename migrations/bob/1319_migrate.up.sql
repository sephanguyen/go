ALTER TABLE public.lesson_student_subscriptions
	ADD COLUMN package_type TEXT NULL;

DROP VIEW IF EXISTS public.student_course_slot_info;
DROP VIEW IF EXISTS public.student_course_recurring_slot_info;

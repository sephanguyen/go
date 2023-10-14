CREATE INDEX IF NOT EXISTS lesson_student_subscripton__package_type__idx ON public.lesson_student_subscriptions  (package_type);

CREATE INDEX IF NOT EXISTS lessons__class_id__idx ON public.lessons  (class_id);
CREATE INDEX IF NOT EXISTS lesson_members__user_id__idx ON public.lesson_members  (user_id);

CREATE INDEX IF NOT EXISTS class_member__user_id__idx ON public.class_member  (user_id);

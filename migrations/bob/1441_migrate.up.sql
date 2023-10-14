CREATE INDEX IF NOT EXISTS lesson_student_subscriptions__course_id__idx 
ON public.lesson_student_subscriptions USING btree (course_id);
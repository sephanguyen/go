DROP INDEX IF EXISTS idx__notification_student_courses__location_id;
CREATE INDEX idx__notification_student_courses__location_id ON public.notification_student_courses USING btree (location_id);

DROP INDEX IF EXISTS idx__notification_student_courses__course_id;
CREATE INDEX idx__notification_student_courses__course_id ON public.notification_student_courses USING btree (course_id);

DROP INDEX IF EXISTS idx__notification_student_courses__start_at;
CREATE INDEX idx__notification_student_courses__start_at ON public.notification_student_courses USING btree (start_at);

DROP INDEX IF EXISTS idx__notification_student_courses__end_at;
CREATE INDEX idx__notification_student_courses__end_at ON public.notification_student_courses USING btree (end_at);

DROP INDEX IF EXISTS idx__notification_class_members__start_at;
CREATE INDEX idx__notification_class_members__start_at ON public.notification_class_members USING btree (start_at);

DROP INDEX IF EXISTS idx__notification_class_members__end_at;
CREATE INDEX idx__notification_class_members__end_at ON public.notification_class_members USING btree (end_at);

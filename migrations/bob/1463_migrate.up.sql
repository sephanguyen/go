ALTER TABLE notification_student_courses REPLICA IDENTITY DEFAULT;

ALTER TABLE ONLY public.notification_student_courses ADD CONSTRAINT pk__notification_student_courses PRIMARY KEY (student_course_id);

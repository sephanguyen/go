ALTER TABLE ONLY public.student_enrollment_status_history ADD CONSTRAINT pk__student_enrollment_status_history PRIMARY KEY (student_id, location_id, enrollment_status, start_date);

ALTER PUBLICATION debezium_publication ADD TABLE public.student_enrollment_status_history;
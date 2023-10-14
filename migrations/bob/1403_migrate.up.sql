ALTER TABLE ONLY public.student_enrollment_status_history
    ADD COLUMN IF NOT EXISTS order_id text NULL;

ALTER TABLE ONLY public.student_enrollment_status_history
    ADD COLUMN IF NOT EXISTS order_sequence_number integer NULL;
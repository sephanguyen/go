DROP INDEX IF EXISTS staff__staff_id__idx;
CREATE INDEX staff__resource_path__idx ON public.staff USING btree (resource_path);
CREATE INDEX student_enrollment_status_history__deleted_at_idx ON public.student_enrollment_status_history USING btree (deleted_at);
CREATE INDEX student_enrollment_status_history__location_id_idx ON public.student_enrollment_status_history USING btree (location_id);
CREATE INDEX student_enrollment_status_history__start_date_end_date_location ON public.student_enrollment_status_history USING btree (
    start_date, end_date, location_id
    );
CREATE INDEX student_enrollment_status_history__start_date_idx ON public.student_enrollment_status_history USING btree (start_date);
CREATE INDEX student_enrollment_status_history__student_id_idx ON public.student_enrollment_status_history USING btree (student_id);
CREATE INDEX students__created_at__idx_desc ON public.students USING btree (created_at DESC);
CREATE INDEX students__grade_id__idx ON public.students USING btree (grade_id);
CREATE INDEX students_resource_path_idx ON public.students USING btree (resource_path);
CREATE UNIQUE INDEX student_enrollment_status_his_student_id_location_id_enroll_key ON public.student_enrollment_status_history USING btree (
    student_id, location_id, enrollment_status,
    start_date, end_date
    );

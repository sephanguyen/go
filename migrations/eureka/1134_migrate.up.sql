CREATE INDEX IF NOT EXISTS exam_lo_submission_status_idx ON public.exam_lo_submission(status);

CREATE INDEX IF NOT EXISTS exam_lo_submission_student_id_idx ON public.exam_lo_submission(student_id);

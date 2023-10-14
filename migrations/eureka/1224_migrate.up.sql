CREATE INDEX IF NOT EXISTS student_submissions_study_plan_item_id_idx ON public.student_submissions USING btree(study_plan_item_id);

CREATE INDEX IF NOT EXISTS student_submissions_student_submission_id_idx ON public.student_submissions (student_submission_id);
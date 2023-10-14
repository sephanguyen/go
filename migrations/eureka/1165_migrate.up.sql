CREATE INDEX IF NOT EXISTS student_submissions_study_plan_id_idx ON public.student_submissions(study_plan_id);
CREATE INDEX IF NOT EXISTS student_submissions_learning_material_id_idx ON public.student_submissions(learning_material_id);
CREATE INDEX IF NOT EXISTS student_submissions_student_id_idx ON public.student_submissions(student_id);
CREATE INDEX IF NOT EXISTS student_submissions_study_plan_item_identity_idx ON public.student_submissions(study_plan_id, learning_material_id, student_id);

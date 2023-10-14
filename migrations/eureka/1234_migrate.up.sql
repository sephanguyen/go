CREATE INDEX CONCURRENTLY IF NOT EXISTS exam_lo_submission_study_plan_item_identity_idx ON public.exam_lo_submission(study_plan_id, student_id, learning_material_id);

CREATE INDEX IF NOT EXISTS shuffled_quiz_sets_study_plan_item_identity_idx ON public.shuffled_quiz_sets(student_id, study_plan_id, learning_material_id);

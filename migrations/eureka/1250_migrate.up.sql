CREATE UNIQUE INDEX lo_progression_study_plan_item_identity_un ON public.lo_progression(student_id, study_plan_id, learning_material_id) WHERE (deleted_at IS NULL);

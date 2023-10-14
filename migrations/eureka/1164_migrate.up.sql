CREATE INDEX CONCURRENTLY IF NOT EXISTS student_event_logs_study_plan_item_identity_idx ON public.student_event_logs(study_plan_id, learning_material_id, student_id);

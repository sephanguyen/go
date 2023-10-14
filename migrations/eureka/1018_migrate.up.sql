-- ALTER TABLE lo_study_plan_items DROP CONSTRAINT lo_study_plan_items_pk;
-- ALTER TABLE	lo_study_plan_items ADD CONSTRAINT lo_study_plan_items_pk PRIMARY KEY (study_plan_item_id, lo_id);

-- ALTER TABLE assignment_study_plan_items DROP CONSTRAINT assignment_study_plan_items_pk;
-- ALTER TABLE	assignment_study_plan_items ADD CONSTRAINT assignment_study_plan_items_pk PRIMARY KEY (study_plan_item_id, assignment_id);

CREATE INDEX IF NOT EXISTS student_study_plans_student_id_idx ON public.student_study_plans (student_id);

CREATE INDEX IF NOT EXISTS study_plan_items_study_plan_id_idx ON public.study_plan_items (study_plan_id);
ALTER TABLE public.lms_student_study_plan_item
    ADD COLUMN IF NOT EXISTS study_plan_item_id text NULL;

ALTER TABLE public.lms_student_study_plan_item
    ADD COLUMN IF NOT EXISTS master_study_plan_item_id text NULL;

ALTER TABLE public.lms_student_study_plan_item
    DROP CONSTRAINT IF EXISTS lms_student_study_plan_item_pkey,
    ADD CONSTRAINT lms_student_study_plan_item_pkey PRIMARY KEY (study_plan_item_id);

ALTER TABLE public.lms_study_plans
    ALTER COLUMN academic_year type text;

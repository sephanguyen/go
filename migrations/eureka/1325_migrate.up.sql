-- Remove un-valid assessment_session records
DELETE FROM public.assessment_session WHERE assessment_id IS NULL;

CREATE TABLE IF NOT EXISTS public.study_plan_assessment(
    id TEXT NOT NULL,
    study_plan_item_id TEXT NOT NULL,
    learning_material_id TEXT NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text DEFAULT public.autofillresourcepath(),
    ref_table VARCHAR(20) NOT NULL,

    CONSTRAINT pk__sp_assessment_id PRIMARY KEY (id),
    CONSTRAINT fk__sp_item_id FOREIGN KEY (study_plan_item_id) REFERENCES public.lms_study_plan_items (study_plan_item_id),
    CONSTRAINT fk__learning_material_id FOREIGN KEY (learning_material_id) REFERENCES public.learning_material (learning_material_id),
    CONSTRAINT un_lm_sp_item UNIQUE (learning_material_id, study_plan_item_id)
    );

CREATE POLICY rls_study_plan_assessment ON "study_plan_assessment" using (
    permission_check(resource_path, 'study_plan_assessment')
) with check (
    permission_check(resource_path, 'study_plan_assessment')
);
CREATE POLICY rls_study_plan_assessment_restrictive ON "study_plan_assessment" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'study_plan_assessment')
) with check (
    permission_check(resource_path, 'study_plan_assessment')
);
ALTER TABLE public.study_plan_assessment ENABLE ROW LEVEL security;
ALTER TABLE public.study_plan_assessment FORCE ROW LEVEL security;

-- Alter assessment_session
ALTER TABLE public.assessment_session
    ALTER COLUMN assessment_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS study_plan_assessment_id TEXT,
    ADD CONSTRAINT fk_sp_assessment_id FOREIGN KEY (study_plan_assessment_id) REFERENCES public.study_plan_assessment(id),
    -- Only one of the two columns can be not null
    ADD CONSTRAINT check_one_of_fk_assessment_not_null CHECK (
      CASE WHEN assessment_id IS NULL THEN 0 ELSE 1 END +
      CASE WHEN study_plan_assessment_id  IS NULL THEN 0 ELSE 1 END = 1
    );

-- Alter assessment_submission

ALTER TABLE public.assessment_submission
    ALTER COLUMN assessment_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS study_plan_assessment_id TEXT,
    -- Only one of the two columns can be not null
    ADD CONSTRAINT check_one_of_fk_assessment_not_null CHECK (
      CASE WHEN assessment_id IS NULL THEN 0 ELSE 1 END +
      CASE WHEN study_plan_assessment_id  IS NULL THEN 0 ELSE 1 END = 1
    );

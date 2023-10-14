CREATE TABLE IF NOT EXISTS public.flash_card_submission (
    submission_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    status text,
    result text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    total_point integer DEFAULT 0,

    CONSTRAINT flash_card_submission_pk PRIMARY KEY (submission_id),
    CONSTRAINT flash_card_submission_flash_card_fk FOREIGN KEY (learning_material_id) REFERENCES public.flash_card(learning_material_id),
    CONSTRAINT flash_card_submission_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id),
    CONSTRAINT flash_card_submission_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id)
);

/* set RLS */
CREATE POLICY rls_flash_card_submission ON "flash_card_submission" using (
    permission_check(resource_path, 'flash_card_submission')
) with check (
    permission_check(resource_path, 'flash_card_submission')
);

CREATE POLICY rls_flash_card_submission_restrictive ON "flash_card_submission" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'flash_card_submission')
) with check (
    permission_check(resource_path, 'flash_card_submission')
);

ALTER TABLE IF EXISTS "flash_card_submission" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "flash_card_submission" FORCE ROW LEVEL security;

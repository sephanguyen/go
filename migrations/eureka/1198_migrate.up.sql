CREATE TABLE IF NOT EXISTS public.flash_card_submission_answer (
    student_id text NOT NULL,
    quiz_id text NOT NULL,
    submission_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    student_text_answer text[],
    correct_text_answer text[],
    student_index_answer integer[],
    correct_index_answer integer[],
    is_correct boolean[] DEFAULT '{}'::boolean[] NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    point integer DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT flash_card_submission_answer_pk PRIMARY KEY (student_id, quiz_id, submission_id),
    CONSTRAINT flash_card_submission_answer_flash_card_submission_fk FOREIGN KEY (submission_id)  REFERENCES public.flash_card_submission(submission_id)
);

/* set RLS */
CREATE POLICY rls_flash_card_submission_answer ON "flash_card_submission_answer" using (
    permission_check(resource_path, 'flash_card_submission_answer')
) with check (
    permission_check(resource_path, 'flash_card_submission_answer')
);

CREATE POLICY rls_flash_card_submission_answer_restrictive ON "flash_card_submission_answer" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'flash_card_submission_answer')
) with check (
    permission_check(resource_path, 'flash_card_submission_answer')
);

ALTER TABLE IF EXISTS "flash_card_submission_answer" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "flash_card_submission_answer" FORCE ROW LEVEL security;
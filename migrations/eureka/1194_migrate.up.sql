CREATE TABLE IF NOT EXISTS public.lo_submission_answer (
  -- PRIMARY KEY
    student_id text NOT NULL,
    quiz_id text NOT NULL,
    submission_id text NOT NULL,
  -- FOREIGN KEY (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
  -- FIELDS
    student_TEXT_answer TEXT[],
    correct_TEXT_answer TEXT[],
    student_index_answer INTEGER[],
    correct_index_answer INTEGER[],
    point INTEGER DEFAULT 0,
    is_correct BOOLEAN[] DEFAULT '{}'::BOOLEAN[] NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT public.autofillresourcepath(),
    CONSTRAINT lo_submission_answer_pk PRIMARY KEY (student_id, quiz_id, submission_id),
    CONSTRAINT lo_submission_answer_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id),
    CONSTRAINT lo_submission_answer_learning_objective_fk FOREIGN KEY (learning_material_id) REFERENCES public.learning_objective(learning_material_id),
    CONSTRAINT lo_submission_answer_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id)
);

/* set RLS */
CREATE POLICY rls_lo_submission_answer ON "lo_submission_answer" using (
  permission_check(resource_path, 'lo_submission_answer')
) with check (
  permission_check(resource_path, 'lo_submission_answer')
);

CREATE POLICY rls_lo_submission_answer_restrictive ON "lo_submission_answer" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lo_submission_answer')
) with check (
    permission_check(resource_path, 'lo_submission_answer')
);

ALTER TABLE IF EXISTS "lo_submission_answer" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "lo_submission_answer" FORCE ROW LEVEL security;
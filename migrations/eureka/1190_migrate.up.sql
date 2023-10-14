CREATE TABLE IF NOT EXISTS public.lo_submission (
    submission_id text NOT NULL PRIMARY KEY,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    status text,
    result text,
    total_point integer DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath(),
    
    CONSTRAINT lo_submission_learning_objective_fk FOREIGN KEY (learning_material_id) REFERENCES public.learning_objective(learning_material_id),
    CONSTRAINT lo_submission_shuffled_quiz_sets_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id),
    CONSTRAINT lo_submission_students_fk FOREIGN KEY (shuffled_quiz_set_id) REFERENCES public.shuffled_quiz_sets(shuffled_quiz_set_id),
    CONSTRAINT lo_submission_study_plans_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans(study_plan_id)
);

/* set RLS */
CREATE POLICY rls_lo_submission ON "lo_submission" using (
    permission_check(resource_path, 'lo_submission')
) with check (
    permission_check(resource_path, 'lo_submission')
);

CREATE POLICY rls_lo_submission_restrictive ON "lo_submission" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lo_submission')
) with check (
    permission_check(resource_path, 'lo_submission')
);

ALTER TABLE IF EXISTS "lo_submission" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "lo_submission" FORCE ROW LEVEL security;
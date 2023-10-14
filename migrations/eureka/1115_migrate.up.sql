CREATE TABLE IF NOT EXISTS public.exam_lo_submission (
    submission_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    shuffled_quiz_set_id text NOT NULL,
    status text,
    result text,
    teacher_feedback text,
    teacher_id text,
    marked_at timestamp with time zone,
    removed_at timestamp with time zone,
    total_point integer NOT NULL DEFAULT 1,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    
    CONSTRAINT exam_lo_submission_pk PRIMARY KEY (submission_id)
);

/* set RLS */
CREATE POLICY rls_exam_lo_submission ON "exam_lo_submission" using (
    permission_check(resource_path, 'exam_lo_submission')
) with check (
    permission_check(resource_path, 'exam_lo_submission')
);

CREATE POLICY rls_exam_lo_submission_restrictive ON "exam_lo_submission" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'exam_lo_submission')
) with check (
    permission_check(resource_path, 'exam_lo_submission')
);

ALTER TABLE "exam_lo_submission" ENABLE ROW LEVEL security;
ALTER TABLE "exam_lo_submission" FORCE ROW LEVEL security;
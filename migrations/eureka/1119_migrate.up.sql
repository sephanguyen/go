CREATE TABLE IF NOT EXISTS public.exam_lo_submission_score (
    submission_id text NOT NULL,
    quiz_id text NOT NULL,
    teacher_id text NOT NULL,
    shuffle_quiz_set_id text NOT NULL,
    teacher_comment text,
    is_correct boolean[] NOT NULL DEFAULT '{}'::boolean[],
    is_accepted boolean,
    point integer NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    
    CONSTRAINT exam_lo_submission_score_pk PRIMARY KEY (submission_id, quiz_id),
    CONSTRAINT exam_lo_submission_score_fk FOREIGN KEY (submission_id) REFERENCES public.exam_lo_submission(submission_id)
);

/* set RLS */
CREATE POLICY rls_exam_lo_submission_score ON "exam_lo_submission_score" using (
    permission_check(resource_path, 'exam_lo_submission_score')
) with check (
    permission_check(resource_path, 'exam_lo_submission_score')
);

CREATE POLICY rls_exam_lo_submission_score_restrictive ON "exam_lo_submission_score" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'exam_lo_submission_score')
) with check (
    permission_check(resource_path, 'exam_lo_submission_score')
);

ALTER TABLE "exam_lo_submission_score" ENABLE ROW LEVEL security;
ALTER TABLE "exam_lo_submission_score" FORCE ROW LEVEL security;
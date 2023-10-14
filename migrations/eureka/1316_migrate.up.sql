ALTER TABLE public.feedback_session
    DROP CONSTRAINT IF EXISTS feedback_session_fk;

ALTER TABLE public.feedback_session
    DROP CONSTRAINT student_session_id_un;

ALTER TABLE public.feedback_session
    RENAME COLUMN student_session_id TO submission_id;

ALTER TABLE public.feedback_session
    ADD CONSTRAINT feedback_session_fk FOREIGN KEY (submission_id) REFERENCES public.assessment_submission(id);

ALTER TABLE public.feedback_session
    ADD CONSTRAINT submission_id_un UNIQUE (submission_id);

ALTER TABLE public.assessment_submission
    ADD COLUMN IF NOT EXISTS completed_at timestamptz NOT NULL;

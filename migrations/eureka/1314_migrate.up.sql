ALTER TABLE public.assessment_submission
    DROP CONSTRAINT IF EXISTS status_check;

ALTER TABLE public.assessment_submission
    RENAME COLUMN status TO grading_status;

ALTER TABLE public.assessment_submission
    ADD CONSTRAINT grading_status_check CHECK ((grading_status = ANY (ARRAY['NOT_MARKED', 'IN_PROGRESS', 'MARKED', 'RETURNED'])));

ALTER TABLE public.assessment_submission
    ADD CONSTRAINT session_id_un UNIQUE (session_id);

ALTER TABLE public.assessment_submission
    RENAME COLUMN total_score TO max_score;

ALTER TABLE public.assessment_submission
    RENAME COLUMN total_gained_score TO graded_score;

ALTER TABLE public.feedback_session
    ADD CONSTRAINT student_session_id_un UNIQUE (student_session_id);

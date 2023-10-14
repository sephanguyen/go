CREATE TABLE IF NOT EXISTS public.assessment_submission
(
    id text, -- Auto-generated ULID.
    session_id text, -- FK to assessment_session.session_id
    assessment_id text not null, -- assessment_session.assessment_id
    student_id text not null, -- assessment_session.user_id
    status text not null, -- Not Marked, In Progress, Marked, Returned.
    total_score integer not null DEFAULT 0, -- Total score is stored in case Item is updated.
    total_gained_score integer not null DEFAULT 0, -- Total grained score of a student.
    allocated_marker_id text, -- who is assigned.
    marked_by text, -- who changes the status to Marked.
    marked_at timestamptz, -- which time the status is changed to Marked.

    created_at timestamptz not null,
    updated_at timestamptz not null,
    deleted_at timestamptz,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT assessment_submission_pk PRIMARY KEY (id),
    CONSTRAINT assessment_submission_fk FOREIGN KEY (session_id) REFERENCES public.assessment_session(session_id),
    CONSTRAINT status_check CHECK ((status = ANY (ARRAY['NOT_MARKED', 'IN_PROGRESS', 'MARKED', 'RETURNED'])))
);

/* set RLS */
CREATE POLICY rls_assessment_submission ON "assessment_submission" using (
    permission_check(resource_path, 'assessment_submission')
    ) with check (
    permission_check(resource_path, 'assessment_submission')
    );

CREATE POLICY rls_assessment_submission_restrictive ON "assessment_submission" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'assessment_submission')
    ) with check (
    permission_check(resource_path, 'assessment_submission')
    );

ALTER TABLE "assessment_submission" ENABLE ROW LEVEL security;
ALTER TABLE "assessment_submission" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.feedback_session
(
    id text, -- Teacher feedback session ID.
    student_session_id text, -- Student session ID.

    created_by text not null, -- a first teacher goes to submission detail page.
    created_at timestamptz not null,
    updated_at timestamptz not null,
    deleted_at timestamptz,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT feedback_session_pk PRIMARY KEY (id),
    CONSTRAINT feedback_session_fk FOREIGN KEY (student_session_id) REFERENCES public.assessment_session(session_id)
);

/* set RLS */
CREATE POLICY rls_feedback_session ON "feedback_session" using (
    permission_check(resource_path, 'feedback_session')
    ) with check (
    permission_check(resource_path, 'feedback_session')
    );

CREATE POLICY rls_feedback_session_restrictive ON "feedback_session" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'feedback_session')
    ) with check (
    permission_check(resource_path, 'feedback_session')
    );

ALTER TABLE "feedback_session" ENABLE ROW LEVEL security;
ALTER TABLE "feedback_session" FORCE ROW LEVEL security;

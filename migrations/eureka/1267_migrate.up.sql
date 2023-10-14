DROP TABLE IF EXISTS public.learnosity_lo_session;

-- rename from learnosity_lo_session to assessment_session
CREATE TABLE IF NOT EXISTS public.assessment_session (
    session_id text NOT NULL,
    learning_material_id text NOT NULL,
    study_plan_id text NOT NULL, -- master_study_plan_id
    user_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    
    CONSTRAINT assessment_session_pk PRIMARY KEY (session_id)
);

/* set RLS */
CREATE POLICY rls_assessment_session ON "assessment_session" using (
    permission_check(resource_path, 'assessment_session')
) with check (
    permission_check(resource_path, 'assessment_session')
);

CREATE POLICY rls_assessment_session_restrictive ON "assessment_session" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'assessment_session')
) with check (
    permission_check(resource_path, 'assessment_session')
);

ALTER TABLE "assessment_session" ENABLE ROW LEVEL security;
ALTER TABLE "assessment_session" FORCE ROW LEVEL security;

-- Create index for searching latest session by identity
CREATE INDEX assessment_session_latest_session_by_identity_idx
    ON assessment_session (learning_material_id, study_plan_id, user_id, created_at DESC);
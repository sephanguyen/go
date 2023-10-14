CREATE TABLE IF NOT EXISTS public.learnosity_lo_session (
    session_id text NOT NULL,
    learning_material_id text NOT NULL,
    study_plan_id text NOT NULL, -- master_study_plan_id
    user_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    
    CONSTRAINT learnosity_lo_session_pk PRIMARY KEY (session_id)
);

/* set RLS */
CREATE POLICY rls_learnosity_lo_session ON "learnosity_lo_session" using (
    permission_check(resource_path, 'learnosity_lo_session')
) with check (
    permission_check(resource_path, 'learnosity_lo_session')
);

CREATE POLICY rls_learnosity_lo_session_restrictive ON "learnosity_lo_session" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'learnosity_lo_session')
) with check (
    permission_check(resource_path, 'learnosity_lo_session')
);

ALTER TABLE "learnosity_lo_session" ENABLE ROW LEVEL security;
ALTER TABLE "learnosity_lo_session" FORCE ROW LEVEL security;

-- Create index for searching latest session by identity
CREATE INDEX learnosity_lo_session_latest_session_by_identity_idx
    ON learnosity_lo_session (learning_material_id, study_plan_id, user_id, created_at DESC);
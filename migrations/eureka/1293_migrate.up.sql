ALTER TABLE public.assessment_session
    ADD COLUMN IF NOT EXISTS course_id text;

-- Create index for searching latest session by identity
CREATE INDEX assessment_session_latest_session_by_identity_v2_idx
    ON assessment_session (learning_material_id, user_id, course_id, created_at DESC);

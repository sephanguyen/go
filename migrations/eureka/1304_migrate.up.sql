DROP INDEX IF EXISTS assessment_session_latest_session_by_identity_idx;
DROP INDEX IF EXISTS assessment_session_latest_session_by_identity_v2_idx;

ALTER TABLE public.assessment_session DROP COLUMN course_id;

ALTER TABLE public.assessment_session ALTER COLUMN learning_material_id DROP NOT NULL;

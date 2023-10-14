-- SPEC: https://manabie.atlassian.net/wiki/spaces/TECH/pages/715030580/LMS+2.0+Tech+specs+Get+Learning+Objective+statuses
CREATE TABLE IF NOT EXISTS public.assessment(
    id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    learning_material_id TEXT NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text DEFAULT public.autofillresourcepath(),

    CONSTRAINT pk__assessment_id PRIMARY KEY (id),
    CONSTRAINT fk__course_id FOREIGN KEY (course_id) REFERENCES public.courses (course_id),
    CONSTRAINT fk__learning_material_id FOREIGN KEY (learning_material_id) REFERENCES public.learning_material (learning_material_id),
    CONSTRAINT un_lm_course UNIQUE (learning_material_id, course_id)
    );

CREATE POLICY rls_assessment ON "assessment" using (
    permission_check(resource_path, 'assessment')
) with check (
    permission_check(resource_path, 'assessment')
);
CREATE POLICY rls_assessment_restrictive ON "assessment" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'assessment')
) with check (
    permission_check(resource_path, 'assessment')
);
ALTER TABLE public.assessment ENABLE ROW LEVEL security;
ALTER TABLE public.assessment FORCE ROW LEVEL security;

ALTER TABLE public.assessment_session
    ADD COLUMN IF NOT EXISTS assessment_id TEXT,
    ADD CONSTRAINT fk_assessment_assessment_id FOREIGN KEY (assessment_id) REFERENCES public.assessment(id);
-- Update it to the FK not null in the next sprint

CREATE INDEX latest_session_by_identity_idx
    ON assessment_session (assessment_id, user_id, created_at DESC);

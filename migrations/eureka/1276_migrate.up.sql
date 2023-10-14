CREATE TABLE IF NOT EXISTS public.max_score_submission (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    student_id text NOT NULL,
    max_score integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT max_score_submission_study_plan_item_identity_pk PRIMARY KEY (learning_material_id, student_id, study_plan_id)
);

/* set RLS */
CREATE POLICY rls_max_score_submission ON "max_score_submission" using (
  permission_check(resource_path, 'max_score_submission')
) with check (
  permission_check(resource_path, 'max_score_submission')
);

ALTER TABLE "max_score_submission" ENABLE ROW LEVEL security;
ALTER TABLE "max_score_submission" FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS user_group_member_user_group_idx ON public.user_group_member(user_group_id);
CREATE TABLE IF NOT EXISTS public.question_group (
    question_group_id    TEXT NOT NULL,
    learning_material_id TEXT NOT NULL,
    name          TEXT,
    description   TEXT,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);

CREATE POLICY rls_question_group ON "question_group" USING (permission_check(resource_path, 'question_group')) WITH CHECK (permission_check(resource_path, 'question_group'));

CREATE POLICY rls_question_group_restrictive ON "question_group"  AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'question_group'))
WITH CHECK (permission_check(resource_path, 'question_group'));

ALTER TABLE IF EXISTS public.question_group ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.question_group FORCE ROW LEVEL security;

ALTER TABLE IF EXISTS public.quiz_sets ADD COLUMN question_hierachy JSON;

ALTER TABLE IF EXISTS quizzes ADD COLUMN question_group_id TEXT NULL;

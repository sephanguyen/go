CREATE TABLE IF NOT EXISTS public.question_tag_type (
    question_tag_type_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT question_tag_type_id_pk PRIMARY KEY (question_tag_type_id)
);

/* set RLS */
CREATE POLICY rls_question_tag_type ON "question_tag_type" 
USING (permission_check(resource_path, 'question_tag_type')) 
WITH CHECK (permission_check(resource_path, 'question_tag_type'));

CREATE POLICY rls_question_tag_type_restrictive ON "question_tag_type"  AS RESTRICTIVE TO PUBLIC 
USING (permission_check(resource_path, 'question_tag_type'))
WITH CHECK (permission_check(resource_path, 'question_tag_type'));

ALTER TABLE IF EXISTS "question_tag_type" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "question_tag_type" FORCE ROW LEVEL security;


-- Add table question_tag
CREATE TABLE IF NOT EXISTS public.question_tag (
    question_tag_id text NOT NULL,
    name text NOT NULL,
    question_tag_type_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT question_tag_type_id_fk FOREIGN KEY (question_tag_type_id) REFERENCES public.question_tag_type(question_tag_type_id),
    CONSTRAINT question_tag_id_pk PRIMARY KEY (question_tag_id)
);

/* set RLS */
CREATE POLICY rls_question_tag ON "question_tag" 
USING (permission_check(resource_path, 'question_tag')) 
WITH CHECK (permission_check(resource_path, 'question_tag'));

CREATE POLICY rls_question_tag_restrictive ON "question_tag"  AS RESTRICTIVE TO PUBLIC 
USING (permission_check(resource_path, 'question_tag'))
WITH CHECK (permission_check(resource_path, 'question_tag'));

ALTER TABLE IF EXISTS "question_tag" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "question_tag" FORCE ROW LEVEL security;

-- add column question_tag_ids to table quizzes
ALTER TABLE IF EXISTS public.quizzes 
    ADD COLUMN IF NOT EXISTS question_tag_ids _text NULL;
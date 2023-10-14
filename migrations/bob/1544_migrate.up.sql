CREATE TABLE IF NOT EXISTS public.questionnaire_templates (
    questionnaire_template_id text NOT NULL,
    "name" text NOT NULL,
    resubmit_allowed bool NOT NULL DEFAULT false,
    expiration_date timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    "type" text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__questionnaire_templates PRIMARY KEY (questionnaire_template_id),
    CONSTRAINT questionnaire_templates_type__check CHECK ((type = ANY (ARRAY['QUESTION_TML_TYPE_DEFAULT'::text])))
);

CREATE POLICY rls_questionnaire_templates ON "questionnaire_templates" 
using (permission_check(resource_path, 'questionnaire_templates')) 
with check (permission_check(resource_path, 'questionnaire_templates'));

CREATE POLICY rls_questionnaire_templates_restrictive ON "questionnaire_templates" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'questionnaire_templates'))
with check (permission_check(resource_path, 'questionnaire_templates'));

ALTER TABLE "questionnaire_templates" ENABLE ROW LEVEL security;
ALTER TABLE "questionnaire_templates" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.questionnaire_template_questions (
    questionnaire_template_question_id text NOT NULL,
    questionnaire_template_id text NOT NULL,
    order_index int4 NOT NULL,
    "type" text NOT NULL,
    title text NOT NULL,
    choices text[],
    is_required bool NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__questionnaire_template_questions PRIMARY KEY (questionnaire_template_question_id),
    CONSTRAINT fk__questionnaire_template_questions__questionnaire_template_id FOREIGN KEY (questionnaire_template_id) REFERENCES public.questionnaire_templates(questionnaire_template_id),
    CONSTRAINT questionnaire_template_questions_type__check CHECK ((type = ANY (ARRAY['QUESTION_TYPE_MULTIPLE_CHOICE'::text, 'QUESTION_TYPE_CHECK_BOX'::text, 'QUESTION_TYPE_FREE_TEXT'::text])))
);


CREATE POLICY rls_questionnaire_template_questions ON "questionnaire_template_questions" 
using (permission_check(resource_path, 'questionnaire_template_questions')) 
with check (permission_check(resource_path, 'questionnaire_template_questions'));

CREATE POLICY rls_questionnaire_template_questions_restrictive ON "questionnaire_template_questions" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'questionnaire_template_questions'))
with check (permission_check(resource_path, 'questionnaire_template_questions'));

ALTER TABLE "questionnaire_template_questions" ENABLE ROW LEVEL security;
ALTER TABLE "questionnaire_template_questions" FORCE ROW LEVEL security;

ALTER TABLE public.questionnaires 
    ADD COLUMN IF NOT EXISTS questionnaire_template_id text,
    ADD CONSTRAINT fk__questionnaires___questionnaire_template_id FOREIGN KEY (questionnaire_template_id) REFERENCES public.questionnaire_templates(questionnaire_template_id);

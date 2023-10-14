ALTER TABLE IF EXISTS public.questionnaire_templates
    DROP CONSTRAINT IF EXISTS questionnaire_templates_type__check,
    ADD CONSTRAINT questionnaire_templates_type__check CHECK ((type = ANY (ARRAY['QUESTION_TEMPLATE_TYPE_DEFAULT'::text])));

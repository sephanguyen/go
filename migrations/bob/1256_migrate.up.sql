ALTER TABLE IF EXISTS public.questionnaire_questions 
ALTER COLUMN title SET NOT NULL;

ALTER TABLE IF EXISTS public.questionnaire_questions 
ALTER COLUMN is_required SET NOT NULL;

ALTER TABLE IF EXISTS public.questionnaire_questions 
ALTER COLUMN is_required SET DEFAULT FALSE;

ALTER TABLE IF EXISTS public.questionnaires 
ALTER COLUMN resubmit_allowed SET NOT NULL;

ALTER TABLE IF EXISTS public.questionnaires 
ALTER COLUMN resubmit_allowed SET DEFAULT FALSE;


CREATE TABLE IF NOT EXISTS public.questionnaires
(
    questionnaire_id text NOT NULL,
    resubmit_allowed bool,
    end_date timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),

	CONSTRAINT pk__questionnaires PRIMARY KEY (questionnaire_id)
);

CREATE TABLE IF NOT EXISTS public.questionnaire_questions
(
	questionnaire_question_id text NOT NULL,
    questionnaire_id text NOT NULL,
	order_index int NOT NULL,
	type text NOT NULL,
	title text,
	choices text[],
    is_required bool,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),

	CONSTRAINT pk__questionnaire_questions PRIMARY KEY (questionnaire_question_id),
	CONSTRAINT fk__questionnaire_questions__questionnaire_id FOREIGN KEY(questionnaire_id) REFERENCES public.questionnaires(questionnaire_id)
);

ALTER TABLE ONLY public.questionnaire_questions DROP CONSTRAINT IF EXISTS questionnaire__question_type__check;
ALTER TABLE public.questionnaire_questions
    ADD CONSTRAINT questionnaire__question_type__check CHECK (type = ANY (ARRAY[
		'QUESTION_TYPE_MULTIPLE_CHOICE',
		'QUESTION_TYPE_CHECK_BOX',
		'QUESTION_TYPE_FREE_TEXT'
]::text[]));

ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS is_important bool;
ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS questionnaire_id text;

ALTER TABLE ONLY public.info_notifications DROP CONSTRAINT IF EXISTS fk__info_notification__questionnaire_id;
ALTER TABLE public.info_notifications ADD CONSTRAINT fk__info_notification__questionnaire_id FOREIGN KEY(questionnaire_id) REFERENCES public.questionnaires(questionnaire_id);

CREATE TABLE IF NOT EXISTS public.questionnaire_user_answers
(
    answer_id text NOT NULL,
	user_notification_id text NOT NULL,
	questionnaire_question_id text NOT NULL,
	user_id text NOT NULL,
	target_id text NOT NULL,
	answer text COLLATE pg_catalog."default",
    submitted_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),

	CONSTRAINT pk__questionnaire_user_answers PRIMARY KEY (answer_id, user_notification_id),
	CONSTRAINT fk__questionnaire_user_answers__user_notification_id FOREIGN KEY(user_notification_id) REFERENCES public.users_info_notifications(user_notification_id),
	CONSTRAINT fk__questionnaire_user_answers__questionnaire_question_id FOREIGN KEY(questionnaire_question_id) REFERENCES public.questionnaire_questions(questionnaire_question_id)
);

ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS is_important bool;

ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS qn_status text DEFAULT 'USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED';

ALTER TABLE ONLY public.users_info_notifications DROP CONSTRAINT IF EXISTS users_info_notifications__qn_status__check;
ALTER TABLE public.users_info_notifications
    ADD CONSTRAINT users_info_notifications__qn_status__check CHECK (qn_status = ANY (ARRAY[
		'USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED',
		'USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED'
]::text[]));


CREATE POLICY rls_questionnaires ON "questionnaires" 
using (permission_check(resource_path, 'questionnaires')) 
with check (permission_check(resource_path, 'questionnaires'));

ALTER TABLE "questionnaires" ENABLE ROW LEVEL security;
ALTER TABLE "questionnaires" FORCE ROW LEVEL security;

CREATE POLICY rls_questionnaire_questions ON "questionnaire_questions" 
using (permission_check(resource_path, 'questionnaire_questions')) 
with check (permission_check(resource_path, 'questionnaire_questions'));

ALTER TABLE "questionnaire_questions" ENABLE ROW LEVEL security;
ALTER TABLE "questionnaire_questions" FORCE ROW LEVEL security;

CREATE POLICY rls_questionnaire_user_answers ON "questionnaire_user_answers" 
using (permission_check(resource_path, 'questionnaire_user_answers')) 
with check (permission_check(resource_path, 'questionnaire_user_answers'));

ALTER TABLE "questionnaire_user_answers" ENABLE ROW LEVEL security;
ALTER TABLE "questionnaire_user_answers" FORCE ROW LEVEL security;

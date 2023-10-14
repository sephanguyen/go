CREATE TABLE IF NOT EXISTS public.conversation_events (
    conversation_event_id text,
    conversation_id text,
    pool_id text,
	user_id text,
	event_type text,
    status text,
    duration int,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS public.pools (
    pool_id text,
    conversation_id text,
    payload jsonb,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL
);


ALTER TABLE ONLY public.conversation_events DROP CONSTRAINT IF EXISTS conversation_events_pk;
ALTER TABLE ONLY public.conversation_events
    ADD CONSTRAINT conversation_events_pk PRIMARY KEY (conversation_event_id);

ALTER TABLE ONLY public.pools DROP CONSTRAINT IF EXISTS pool_id_pk;
ALTER TABLE ONLY public.pools
    ADD CONSTRAINT pool_id_pk PRIMARY KEY (pool_id);

ALTER TABLE ONLY public.conversation_events DROP CONSTRAINT IF EXISTS event_type_check;

ALTER TABLE public.conversation_events
    ADD CONSTRAINT event_type_check CHECK (event_type = ANY (ARRAY[
                'CONVERSATION_EVENT_TYPE_STUDENT_HAND_EVENT'::text,
                'CONVERSATION_EVENT_TYPE_ALLOW_PROHIBIT_TO_SPEAK'::text,
                'CONVERSATION_EVENT_TYPE_TEACHER_CREATE_POOL'::text,
                'CONVERSATION_EVENT_TYPE_STUDENT_ANSWER_POOL'::text
            ]));

ALTER TABLE ONLY public.conversation_events DROP CONSTRAINT IF EXISTS status_check;


ALTER TABLE public.conversation_events
    ADD CONSTRAINT status_check CHECK (status = ANY (ARRAY[
                'CONVERSATION_EVENT_STATUS_STUDENT_RAISE_HAND'::text,
                'CONVERSATION_EVENT_STATUS_STUDENT_PUT_HAND_DOWN'::text,
                'CONVERSATION_EVENT_STATUS_POOL_OPEN'::text,
                'CONVERSATION_EVENT_STATUS_POOL_CLOSE'::text
            ]));

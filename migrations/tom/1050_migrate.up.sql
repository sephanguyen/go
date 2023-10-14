-- New "conversation" table
CREATE TABLE IF NOT EXISTS public.conversation (
    conversation_id text NOT NULL,
    name text,
    latest_message jsonb,
    latest_message_sent_time timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,
    optional_config jsonb,

    CONSTRAINT conversation_pk PRIMARY KEY (conversation_id)
);

CREATE POLICY rls_conversation ON "conversation" using (
    permission_check(resource_path, 'conversation')
) with check (
    permission_check(resource_path, 'conversation')
);

CREATE POLICY rls_conversation_restrictive ON "conversation" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'conversation')
) with check (
    permission_check(resource_path, 'conversation')
);

ALTER TABLE "conversation" ENABLE ROW LEVEL security;
ALTER TABLE "conversation" FORCE ROW LEVEL security;


-- New "conversation_member" table
CREATE TABLE IF NOT EXISTS public.conversation_member (
    conversation_member_id text NOT NULL,
    conversation_id text NOT NULL,
    user_id text NOT NULL,
    status text NOT NULL,
    seen_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT conversation_member_pk PRIMARY KEY (conversation_member_id),
    CONSTRAINT conversation_member_conversation_fk FOREIGN KEY (conversation_id) REFERENCES public.conversation(conversation_id),
    CONSTRAINT conversation_member_conversation_id_user_id_un UNIQUE (conversation_id, user_id)
);

CREATE POLICY rls_conversation_member ON "conversation_member" using (
    permission_check(resource_path, 'conversation_member')
) with check (
    permission_check(resource_path, 'conversation_member')
);

CREATE POLICY rls_conversation_member_restrictive ON "conversation_member" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'conversation_member')
) with check (
    permission_check(resource_path, 'conversation_member')
);

ALTER TABLE "conversation_member" ENABLE ROW LEVEL security;
ALTER TABLE "conversation_member" FORCE ROW LEVEL security;

ALTER TABLE public.conversation_member ADD CONSTRAINT conversation_member_status_check CHECK ((status = ANY ('{
    CONVERSATION_MEMBER_STATUS_ACTIVE,
    CONVERSATION_MEMBER_STATUS_INACTIVE
}'::text[])));

-- New "message" table
CREATE TABLE IF NOT EXISTS public.message (
    message_id text NOT NULL,
    conversation_id text NOT NULL,
    type text,
    message text,
    user_id text NOT NULL,
    sent_at timestamp with time zone,
    extra_info jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT message_pk PRIMARY KEY (message_id),
    CONSTRAINT message_conversation_fk FOREIGN KEY (conversation_id) REFERENCES public.conversation(conversation_id)
);

CREATE POLICY rls_message ON "message" using (
    permission_check(resource_path, 'message')
) with check (
    permission_check(resource_path, 'message')
);

CREATE POLICY rls_message_restrictive ON "message" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'message')
) with check (
    permission_check(resource_path, 'message')
);

ALTER TABLE "message" ENABLE ROW LEVEL security;
ALTER TABLE "message" FORCE ROW LEVEL security;

ALTER TABLE public.message ADD CONSTRAINT message_type_check CHECK ((type = ANY ('{
    MESSAGE_TYPE_TEXT,
    MESSAGE_TYPE_IMAGE,
    MESSAGE_TYPE_VIDEO,
    MESSAGE_TYPE_AUDIO,
    MESSAGE_TYPE_FILE,
    MESSAGE_TYPE_CUSTOM
}'::text[])));

-- New "internal_chat_user" table
CREATE TABLE IF NOT EXISTS public.internal_admin_user (
    user_id text NOT NULL,
    is_system boolean,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT internal_admin_user_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_internal_admin_user ON "internal_admin_user" using (
    permission_check(resource_path, 'internal_admin_user')
) with check (
    permission_check(resource_path, 'internal_admin_user')
);

CREATE POLICY rls_internal_admin_user_restrictive ON "internal_admin_user" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'internal_admin_user')
) with check (
    permission_check(resource_path, 'internal_admin_user')
);

ALTER TABLE "internal_admin_user" ENABLE ROW LEVEL security;
ALTER TABLE "internal_admin_user" FORCE ROW LEVEL security;


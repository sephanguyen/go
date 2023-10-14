-- New "agora_user" table
CREATE TABLE IF NOT EXISTS public.agora_user (
    user_id text NOT NULL,
    agora_user_id text NOT NULL UNIQUE,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT agora_user_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_agora_user ON "agora_user" using (
    permission_check(resource_path, 'agora_user')
) with check (
    permission_check(resource_path, 'agora_user')
);

CREATE POLICY rls_agora_user_restrictive ON "agora_user" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'agora_user')
) with check (
    permission_check(resource_path, 'agora_user')
);

ALTER TABLE "agora_user" ENABLE ROW LEVEL security;
ALTER TABLE "agora_user" FORCE ROW LEVEL security;

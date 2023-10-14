-- New "agora_user" table
CREATE TABLE IF NOT EXISTS public.agora_user_failure (
    user_id text NOT NULL,
    agora_user_id text NOT NULL,
    is_fix bool,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,
    CONSTRAINT agora_user_failure_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_agora_user_failure ON "agora_user_failure" using (
    permission_check(resource_path, 'agora_user_failure')
) with check (
    permission_check(resource_path, 'agora_user_failure')
);

CREATE POLICY rls_agora_user_failure_restrictive ON "agora_user_failure" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'agora_user_failure')
) with check (
    permission_check(resource_path, 'agora_user_failure')
);

ALTER TABLE
    "agora_user_failure" ENABLE ROW LEVEL security;

ALTER TABLE
    "agora_user_failure" FORCE ROW LEVEL security;

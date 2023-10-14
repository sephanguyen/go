CREATE TABLE IF NOT EXISTS public.user_access_paths (
    user_id text NOT NULL,
    location_id text NOT NULL,
    access_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id)
);

CREATE POLICY rls_user_access_paths ON "user_access_paths" using (
    permission_check(resource_path, 'user_access_paths')
) with check (
    permission_check(resource_path, 'user_access_paths')
);

CREATE POLICY rls_user_access_paths_restrictive ON "user_access_paths" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'user_access_paths')
) with check (
    permission_check(resource_path, 'user_access_paths')
);

ALTER TABLE "user_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "user_access_paths" FORCE ROW LEVEL security;

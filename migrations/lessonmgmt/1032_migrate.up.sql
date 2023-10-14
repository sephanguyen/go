CREATE TABLE IF NOT EXISTS public.user_access_paths (
    user_id text NOT NULL,
    location_id text NOT NULL,
    access_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id)
);
CREATE INDEX IF NOT EXISTS user_access_paths__location_id__idx ON public.user_access_paths USING btree (location_id);
CREATE INDEX IF NOT EXISTS user_access_paths__user_id__idx ON public.user_access_paths USING btree (user_id);

CREATE POLICY rls_user_access_paths ON "user_access_paths" USING (permission_check(resource_path, 'user_access_paths')) WITH CHECK (permission_check(resource_path, 'user_access_paths'));
CREATE POLICY rls_user_access_paths_restrictive ON "user_access_paths" AS RESTRICTIVE FOR ALL TO PUBLIC USING (permission_check(resource_path, 'user_access_paths')) WITH CHECK (permission_check(resource_path, 'user_access_paths'));

ALTER TABLE "user_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "user_access_paths" FORCE ROW LEVEL security;

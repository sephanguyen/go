CREATE TABLE IF NOT EXISTS public.user_access_paths (
    "user_id" text NOT NULL,
    "location_id" text NOT NULL,
    "access_path" text,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),

    CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id),
    CONSTRAINT user_access_paths_users_fk FOREIGN KEY (user_id) REFERENCES "users"(user_id),
    CONSTRAINT user_access_paths_locations_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id)
);

CREATE POLICY rls_user_access_paths ON "user_access_paths" using (permission_check(resource_path, 'user_access_paths')) with check (permission_check(resource_path, 'user_access_paths'));

ALTER TABLE "user_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "user_access_paths" FORCE ROW LEVEL security;
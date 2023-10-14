CREATE TABLE IF NOT EXISTS public.class_member (
    class_member_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    start_date timestamp with time zone default NULL,
    end_date timestamp with time zone default NULL,
    CONSTRAINT pk__class_member PRIMARY KEY (class_member_id)
);

CREATE POLICY rls_class_member ON "class_member" using (permission_check(resource_path, 'class_member')) with check (permission_check(resource_path, 'class_member'));
CREATE POLICY rls_class_member_restrictive ON "class_member" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'class_member')) with check (permission_check(resource_path, 'class_member'));

ALTER TABLE "class_member" ENABLE ROW LEVEL security;
ALTER TABLE "class_member" FORCE ROW LEVEL security; 

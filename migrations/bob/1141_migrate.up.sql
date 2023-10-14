CREATE TABLE IF NOT EXISTS public.subjects (
    subject_id TEXT NOT NULL PRIMARY KEY,
    name text NOT NULL,
    display_name text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT unique__subject_name_resource_path UNIQUE (name, resource_path)
);

CREATE POLICY rls_subjects ON "subjects" using (permission_check(resource_path, 'subjects')) with check (permission_check(resource_path, 'subjects'));

ALTER TABLE "subjects" ENABLE ROW LEVEL security;
ALTER TABLE "subjects" FORCE ROW LEVEL security;
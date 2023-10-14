CREATE TABLE public.school_info (
    school_id TEXT NOT NULL,
    school_name TEXT NOT NULL,
    school_name_phonetic TEXT,
    prefecture TEXT,
    city TEXT,
    is_archived BOOLEAN NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT school_info_pk PRIMARY KEY (school_id)
);

/* set RLS */
CREATE POLICY rls_school_info ON "school_info" using (permission_check(resource_path, 'school_info')) with check (permission_check(resource_path, 'school_info'));

ALTER TABLE "school_info" ENABLE ROW LEVEL security;
ALTER TABLE "school_info" FORCE ROW LEVEL security;

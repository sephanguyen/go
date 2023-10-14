CREATE TABLE IF NOT EXISTS public.students (
    student_id text UNIQUE NOT NULL,
    current_grade smallint,
    school_id integer,
    enrollment_status text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT students_pk PRIMARY KEY (student_id)
);


CREATE POLICY rls_students ON "students" using (permission_check(resource_path, 'students')) with check (permission_check(resource_path, 'students'));

ALTER TABLE "students" ENABLE ROW LEVEL security;
ALTER TABLE "students" FORCE ROW LEVEL security;
CREATE TABLE IF NOT EXISTS public.students
(
    student_id text NOT NULL,
    current_grade smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT student_pk PRIMARY KEY (student_id)
    );

CREATE POLICY rls_students ON "students" using (permission_check(resource_path, 'students')) with check (permission_check(resource_path, 'students'));

ALTER TABLE "students" ENABLE ROW LEVEL security;
ALTER TABLE "students" FORCE ROW LEVEL security;
ALTER TABLE ONLY public.invoice ADD CONSTRAINT invoice_students_fk FOREIGN KEY(student_id) REFERENCES public.students(student_id);
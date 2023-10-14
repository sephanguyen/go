CREATE TABLE IF NOT EXISTS public.student_entryexit_records (
    entryexit_id integer NOT NULL,
    student_id text UNIQUE NOT NULL,
    entry_at timestamp with time zone NOT NULL,
    exit_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT student_entryexit_records_pk PRIMARY KEY (entryexit_id),
    CONSTRAINT student_entryexit_records_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);

--- Set the ID sequence
CREATE SEQUENCE public.student_entryexit_records_id_seq
    AS integer;
    
ALTER SEQUENCE public.student_entryexit_records_id_seq OWNED BY public.student_entryexit_records.entryexit_id;

ALTER TABLE ONLY public.student_entryexit_records ALTER COLUMN entryexit_id SET DEFAULT nextval('public.student_entryexit_records_id_seq'::regclass);


--- Set the permission check policy
CREATE POLICY rls_student_entryexit_records ON "student_entryexit_records" using (permission_check(resource_path, 'student_entryexit_records')) with check (permission_check(resource_path, 'student_entryexit_records'));

ALTER TABLE "student_entryexit_records" ENABLE ROW LEVEL security;
ALTER TABLE "student_entryexit_records" FORCE ROW LEVEL security;

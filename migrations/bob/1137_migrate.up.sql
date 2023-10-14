CREATE TABLE IF NOT EXISTS public.student_qr (
    qr_id integer NOT NULL,
    student_id text UNIQUE NOT NULL,
    qr_url text UNIQUE NOT NULL ,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT student_qr_pk PRIMARY KEY (qr_id),
    CONSTRAINT student_qr_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);

CREATE SEQUENCE public.student_qr_id_seq
    AS integer;
    
ALTER SEQUENCE public.student_qr_id_seq OWNED BY public.student_qr.qr_id;

ALTER TABLE ONLY public.student_qr ALTER COLUMN qr_id SET DEFAULT nextval('public.student_qr_id_seq'::regclass);

CREATE POLICY rls_student_qr ON "student_qr" using (permission_check(resource_path, 'student_qr')) with check (permission_check(resource_path, 'student_qr'));

ALTER TABLE "student_qr" ENABLE ROW LEVEL security;
ALTER TABLE "student_qr" FORCE ROW LEVEL security;

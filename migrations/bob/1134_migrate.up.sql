CREATE TABLE public.grade (
                              id integer NOT NULL,
                              name text NOT NULL,
                              is_archived boolean DEFAULT false NOT NULL,
                              updated_at timestamp with time zone NOT NULL,
                              created_at timestamp with time zone NOT NULL,
                              resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.grade
    ADD CONSTRAINT grade_pk PRIMARY KEY (id);

CREATE SEQUENCE public.grade_id_seq
    AS integer;

ALTER SEQUENCE public.grade_id_seq OWNED BY public.grade.id;

ALTER TABLE ONLY public.grade ALTER COLUMN id SET DEFAULT nextval('public.grade_id_seq'::regclass);

CREATE POLICY rls_grade ON "grade" using (permission_check(resource_path, 'grade')) with check (permission_check(resource_path, 'grade'));

ALTER TABLE "grade" ENABLE ROW LEVEL security;
ALTER TABLE "grade" FORCE ROW LEVEL security;

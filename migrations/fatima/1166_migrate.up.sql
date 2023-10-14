CREATE TABLE IF NOT EXISTS public.student_package_log (
                                                          student_package_log_id integer NOT NULL,
                                                          student_package_id text NOT NULL,
                                                          user_id text NOT NULL,
                                                          action text,
                                                          flow text,
                                                          student_package_object jsonb,
                                                          student_id text,
                                                          course_id text,
                                                          created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT student_package_logs PRIMARY KEY (student_package_log_id),
    CONSTRAINT student_package_log_student_package_fk FOREIGN KEY (student_package_id) REFERENCES "student_packages"(student_package_id),
    CONSTRAINT student_package_log_users_fk FOREIGN KEY (user_id) REFERENCES "users"(user_id)
    );

CREATE SEQUENCE public.student_package_log_id_seq
    AS integer;

ALTER SEQUENCE public.student_package_log_id_seq OWNED BY public.student_package_log.student_package_log_id;

ALTER TABLE ONLY public.student_package_log ALTER COLUMN student_package_log_id
    SET DEFAULT nextval('public.student_package_log_id_seq'::regclass);

CREATE POLICY rls_student_package_log ON "student_package_log" using (
    permission_check(resource_path, 'student_package_log'))
    with check (permission_check(resource_path, 'student_package_log'));

CREATE POLICY rls_student_package_log_restrictive ON "student_package_log"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'student_package_log'))
    WITH CHECK (permission_check(resource_path, 'student_package_log'));

ALTER TABLE public.student_package_log ENABLE ROW LEVEL security;
ALTER TABLE public.student_package_log FORCE ROW LEVEL security;

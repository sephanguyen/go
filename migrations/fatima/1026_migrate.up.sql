CREATE TABLE public.course (
    id text NOT NULL,
    name text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.course ADD CONSTRAINT course_pk PRIMARY KEY (id);

ALTER TABLE ONLY public.product_course ADD CONSTRAINT fk_course_id FOREIGN KEY(course_id) REFERENCES public.course(id);

CREATE POLICY rls_course ON "course" using (permission_check(resource_path, 'course')) with check (permission_check(resource_path, 'course'));

ALTER TABLE "course" ENABLE ROW LEVEL security;
ALTER TABLE "course" FORCE ROW LEVEL security;

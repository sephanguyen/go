CREATE TABLE public.class (
    class_id text NOT NULL,
    course_id text NOT NULL,
    location_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
CONSTRAINT class_id_pk PRIMARY KEY (class_id)
);

CREATE POLICY rls_class ON "class"
    USING (permission_check(resource_path, 'class'))
    WITH CHECK (permission_check(resource_path, 'class'));

CREATE POLICY rls_class_restrictive ON "class"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'class'))
    WITH CHECK (permission_check(resource_path, 'class'));

ALTER TABLE "class" ENABLE ROW LEVEL security;
ALTER TABLE "class" FORCE ROW LEVEL security;


CREATE TABLE public.course_access_paths (
    course_id text NOT NULL,
    location_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id, location_id)
);

CREATE POLICY rls_course_access_paths ON "course_access_paths"
    USING (permission_check(resource_path, 'course_access_paths'))
    WITH CHECK (permission_check(resource_path, 'course_access_paths'));

CREATE POLICY rls_course_access_paths_restrictive ON "course_access_paths"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'course_access_paths'))
    WITH CHECK (permission_check(resource_path, 'course_access_paths'));

ALTER TABLE "course_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "course_access_paths" FORCE ROW LEVEL security;

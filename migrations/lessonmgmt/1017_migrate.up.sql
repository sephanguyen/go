CREATE TABLE IF NOT EXISTS public.class (
    class_id TEXT NOT NULL,
    name TEXT NOT NULL,
    course_id TEXT NOT NULL,
    school_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__class PRIMARY KEY (class_id)
);
CREATE POLICY rls_class ON public.class using (permission_check(resource_path, 'class')) with check (permission_check(resource_path, 'class'));

CREATE POLICY rls_class_restrictive ON "class"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'class'))
    WITH CHECK (permission_check(resource_path, 'class'));

ALTER TABLE public.class ENABLE ROW LEVEL security;
ALTER TABLE public.class FORCE ROW LEVEL security;




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
    CONSTRAINT pk__class PRIMARY KEY (class_id),
    CONSTRAINT fk__class__course_id FOREIGN KEY (course_id) REFERENCES public.courses(course_id),
    CONSTRAINT fk__class__location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);
CREATE POLICY rls_class ON public.class using (permission_check(resource_path, 'class')) with check (permission_check(resource_path, 'class'));

ALTER TABLE public.class ENABLE ROW LEVEL security;
ALTER TABLE public.class FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.class_member (
    class_member_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__class_member PRIMARY KEY (class_member_id),
    CONSTRAINT fk__class_member__class_id FOREIGN KEY (class_id) REFERENCES public.class(class_id)
);

CREATE POLICY rls_class_member ON public.class_member using (permission_check(resource_path, 'class_member')) with check (permission_check(resource_path, 'class_member'));

ALTER TABLE public.class_member ENABLE ROW LEVEL security;
ALTER TABLE public.class_member FORCE ROW LEVEL security; 

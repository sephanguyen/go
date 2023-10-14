CREATE TABLE IF NOT EXISTS public.courses (
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    course_access_paths_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    course_access_paths_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    course_access_paths_deleted_at timestamp with time zone,
    courses_name TEXT NOT NULL,
    grade INT,
    teaching_method TEXT,
    course_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    course_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    course_deleted_at timestamp with time zone,
	CONSTRAINT courses_pk PRIMARY KEY (course_id, location_id)
);

ALTER PUBLICATION kec_publication ADD TABLE public.courses;

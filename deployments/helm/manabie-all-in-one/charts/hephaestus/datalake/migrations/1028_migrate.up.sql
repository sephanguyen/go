CREATE TABLE IF NOT EXISTS fatima.course_access_paths (
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
    CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id, location_id)
);

CREATE TABLE IF NOT EXISTS fatima.courses (
    course_id TEXT NOT NULL,
    name TEXT NOT NULL,
    grade INT,
    teaching_method TEXT,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
    CONSTRAINT courses_pk PRIMARY KEY (course_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
fatima.course_access_paths,
fatima.courses;

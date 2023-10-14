CREATE TABLE IF NOT EXISTS fatima.product_discount (
	discount_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT product_discount_pk PRIMARY KEY (discount_id, product_id)
);

CREATE TABLE IF NOT EXISTS fatima.upcoming_student_course (
	upcoming_student_package_id text NOT NULL,
	student_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    student_package_id TEXT NOT NULL,
    package_type TEXT NULL,
    course_slot int4 NULL,
    course_slot_per_week int4 NULL,
    weight int4 NULL,
    student_start_date timestamp with time zone NOT NULL,
    student_end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
	CONSTRAINT upcoming_student_course_pk PRIMARY KEY (upcoming_student_package_id, student_id, course_id, location_id, student_package_id)
);

CREATE TABLE IF NOT EXISTS fatima.student_course (
    student_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    student_package_id TEXT NOT NULL,
    package_type TEXT NULL,
    course_slot int4 NULL,
    course_slot_per_week int4 NULL,
    weight int4 NULL,
    student_start_date timestamp with time zone NOT NULL,
    student_end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
	CONSTRAINT student_course_pk PRIMARY KEY (student_id, course_id, location_id, student_package_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
fatima.product_discount,
fatima.upcoming_student_course,
fatima.student_course;

CREATE TABLE IF NOT EXISTS public.product_discount (
	discount_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT product_discount_pk PRIMARY KEY (discount_id, product_id)
);

CREATE TABLE IF NOT EXISTS public.upcoming_student_course (
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
	CONSTRAINT upcoming_student_course_pk PRIMARY KEY (upcoming_student_package_id, student_id, course_id, location_id, student_package_id)
);

CREATE TABLE IF NOT EXISTS public.student_course (
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
	CONSTRAINT student_course_pk PRIMARY KEY (student_id, course_id, location_id, student_package_id)
);

ALTER PUBLICATION kec_publication ADD TABLE 
public.product_discount,
public.upcoming_student_course,
public.student_course;

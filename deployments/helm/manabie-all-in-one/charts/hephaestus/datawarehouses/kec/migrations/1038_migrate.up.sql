CREATE TABLE IF NOT EXISTS public.course_academic_year (
	course_id text NOT NULL,
    academic_year_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT course_academic_year_pkey PRIMARY KEY (academic_year_id,course_id)
);

CREATE TABLE IF NOT EXISTS public.course_class (
	course_id text NOT NULL,
    class_id text NOT NULL,
    status text,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT course_class_pkey PRIMARY KEY (class_id,course_id)
);


ALTER PUBLICATION kec_publication ADD TABLE 
    public.course_academic_year,
    public.course_class;

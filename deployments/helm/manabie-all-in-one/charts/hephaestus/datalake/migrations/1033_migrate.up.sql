CREATE TABLE IF NOT EXISTS mastermgmt.course_academic_year (
	course_id text NOT NULL,
    academic_year_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text,

	CONSTRAINT course_academic_year_pkey PRIMARY KEY (academic_year_id,course_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE mastermgmt.course_academic_year;

CREATE TABLE IF NOT EXISTS bob.courses (
	course_id text NOT NULL,
	"name" text NOT NULL,
	subject text NULL,
	grade int2 NULL,
	display_order int2 NULL DEFAULT 0,
	updated_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	created_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	start_date timestamptz NULL,
	end_date timestamptz NULL,
	teacher_ids _text NULL,
	preset_study_plan_id text NULL,
	icon text NULL,
	status text NULL DEFAULT 'COURSE_STATUS_NONE'::text,
	resource_path text,
	teaching_method text NULL,
	course_type_id text NULL,
	remarks text NULL,
	is_archived bool NULL DEFAULT false,
	course_partner_id text NULL,
	CONSTRAINT courses_pk PRIMARY KEY (course_id)
);

CREATE TABLE IF NOT EXISTS bob.course_access_paths (
	course_id text NOT NULL,
	location_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text,
	CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id,location_id)
);


CREATE TABLE IF NOT EXISTS bob.course_type (
	course_type_id text NOT NULL,
	"name" text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text,
	remarks text NULL,
	is_archived bool NULL DEFAULT false,
	CONSTRAINT course_type__pk PRIMARY KEY (course_type_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
bob.course_type, bob.courses, bob.course_access_paths;

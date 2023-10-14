-- bob.users definition

CREATE TABLE IF NOT EXISTS bob.users (
	user_id text NOT NULL,
	country text NOT NULL,
	"name" text NOT NULL,
	avatar text NULL,
	phone_number text NULL,
	email text NULL,
	device_token text NULL,
	allow_notification bool NULL,
	user_group text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	is_tester bool NULL,
	facebook_id text NULL,
	platform text NULL,
	phone_verified bool NULL,
	email_verified bool NULL,
	deleted_at timestamptz NULL,
	given_name text NULL,
	resource_path text NOT NULL,
	last_login_date timestamptz NULL,
	birthday date NULL,
	gender text NULL,
	first_name text NOT NULL DEFAULT ''::text,
	last_name text NOT NULL DEFAULT ''::text,
	first_name_phonetic text NULL,
	last_name_phonetic text NULL,
	full_name_phonetic text NULL,
	remarks text NULL,
	is_system bool NULL DEFAULT false,
	user_external_id text NULL,
	previous_name text NULL,
	CONSTRAINT user_gender_check CHECK ((gender = ANY ('{MALE,FEMALE}'::text[]))),
	CONSTRAINT users__email__key UNIQUE (email, resource_path),
	CONSTRAINT users__facebook_id__key UNIQUE (facebook_id, resource_path),
	CONSTRAINT users_pk PRIMARY KEY (user_id),
	CONSTRAINT users_platform_check CHECK ((platform = ANY ('{PLATFORM_NONE,PLATFORM_IOS,PLATFORM_ANDROID}'::text[]))),
	CONSTRAINT users_user_group_check CHECK ((user_group = ANY (ARRAY['USER_GROUP_STUDENT'::text, 'USER_GROUP_COACH'::text, 'USER_GROUP_TUTOR'::text, 'USER_GROUP_STAFF'::text, 'USER_GROUP_ADMIN'::text, 'USER_GROUP_TEACHER'::text, 'USER_GROUP_PARENT'::text, 'USER_GROUP_CONTENT_ADMIN'::text, 'USER_GROUP_CONTENT_STAFF'::text, 'USER_GROUP_SALES_ADMIN'::text, 'USER_GROUP_SALES_STAFF'::text, 'USER_GROUP_CS_ADMIN'::text, 'USER_GROUP_CS_STAFF'::text, 'USER_GROUP_SCHOOL_ADMIN'::text, 'USER_GROUP_SCHOOL_STAFF'::text, 'USER_GROUP_ORGANIZATION_MANAGER'::text])))
);


-- bob.staff definition

-- Drop table

-- DROP TABLE bob.staff;

CREATE TABLE IF NOT EXISTS bob.staff (
	staff_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	auto_create_timesheet bool NULL DEFAULT false,
	working_status text NOT NULL DEFAULT 'AVAILABLE'::text,
	start_date date NULL,
	end_date date NULL,
	CONSTRAINT pk__staff PRIMARY KEY (staff_id)
);

-- bob.parents definition

CREATE TABLE IF NOT EXISTS bob.parents (
	parent_id text NOT NULL,
	school_id int4 NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT parents__parent_id_pk PRIMARY KEY (parent_id)
);

-- public.students definition

CREATE TABLE IF NOT EXISTS bob.students (
	student_id text NOT NULL,
	current_grade int2 NULL,
	target_university text NULL,
	on_trial bool NOT NULL DEFAULT true,
	billing_date timestamptz NOT NULL,
	birthday date NULL,
	biography text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	total_question_limit int2 NULL DEFAULT 20,
	school_id int4 NULL,
	deleted_at timestamptz NULL,
	additional_data jsonb NULL,
	enrollment_status text NOT NULL DEFAULT 'STUDENT_ENROLLMENT_STATUS_ENROLLED'::text,
	resource_path text NOT NULL,
	student_external_id text NULL,
	student_note text NOT NULL DEFAULT ''::text,
	previous_grade int2 NULL,
	contact_preference text NULL,
	grade_id text NULL,
	CONSTRAINT students_enrollment_status_check CHECK ((enrollment_status = ANY ('{STUDENT_ENROLLMENT_STATUS_POTENTIAL,STUDENT_ENROLLMENT_STATUS_ENROLLED,STUDENT_ENROLLMENT_STATUS_TEMPORARY,STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,STUDENT_ENROLLMENT_STATUS_WITHDRAWN,STUDENT_ENROLLMENT_STATUS_GRADUATED,STUDENT_ENROLLMENT_STATUS_LOA}'::text[]))),
	CONSTRAINT students_pk PRIMARY KEY (student_id)
);

-- public.student_parents definition

CREATE TABLE IF NOT EXISTS bob.student_parents (
	student_id text NOT NULL,
	parent_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	relationship text NOT NULL,
	resource_path text NOT NULL,
	CONSTRAINT student_parents_pk PRIMARY KEY (student_id, parent_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.users;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.staff;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.students;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.parents;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.student_parents;

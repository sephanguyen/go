-- bob.school_level definition

CREATE TABLE IF NOT EXISTS bob.school_level (
	school_level_id text NOT NULL,
	school_level_name text NOT NULL,
	"sequence" int4 NOT NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT school_level__pk PRIMARY KEY (school_level_id),
	CONSTRAINT school_level__sequence__unique UNIQUE (sequence, resource_path)
);

-- bob.grade definition

CREATE TABLE IF NOT EXISTS bob.grade (
	"name" text NOT NULL,
	is_archived bool NOT NULL DEFAULT false,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	resource_path text NOT NULL,
	grade_id text NOT NULL,
	partner_internal_id varchar(50) NOT NULL,
	deleted_at timestamptz NULL,
	"sequence" int4 NULL,
	CONSTRAINT grade_pk PRIMARY KEY (grade_id)
);

-- bob.school_level_grade definition

CREATE TABLE IF NOT EXISTS bob.school_level_grade (
	school_level_id text NOT NULL,
	grade_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT school_level_grade__pk PRIMARY KEY (school_level_id, grade_id)
);

-- bob.school_course definition

CREATE TABLE IF NOT EXISTS bob.school_course (
	school_course_id text NOT NULL,
	school_course_name text NOT NULL,
	school_course_name_phonetic text NULL,
	school_id text NOT NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	school_course_partner_id text NOT NULL,
	CONSTRAINT school_course__pk PRIMARY KEY (school_course_id),
	CONSTRAINT school_course__school_course_partner_id__unique UNIQUE (school_course_partner_id, resource_path)
);

-- bob.school_info definition

CREATE TABLE IF NOT EXISTS bob.school_info (
	school_id text NOT NULL,
	school_name text NOT NULL,
	school_name_phonetic text NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	school_level_id text NOT NULL,
	address text NULL,
	school_partner_id text NOT NULL,
	CONSTRAINT school_info__school_partner_id__unique UNIQUE (school_partner_id, resource_path),
	CONSTRAINT school_info_pk PRIMARY KEY (school_id)
);

-- bob.school_history definition

CREATE TABLE IF NOT EXISTS bob.school_history (
	student_id text NOT NULL,
	school_id text NOT NULL,
	school_course_id text NULL,
	start_date timestamptz NULL,
	end_date timestamptz NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NULL,
	is_current bool NULL DEFAULT false,
	CONSTRAINT school_history__pk PRIMARY KEY (student_id, school_id)
);

-- bob.user_tag definition

CREATE TABLE IF NOT EXISTS bob.user_tag (
	user_tag_id text NOT NULL,
	user_tag_name text NOT NULL,
	user_tag_type text NOT NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	user_tag_partner_id text NOT NULL,
	CONSTRAINT user_tag__pk PRIMARY KEY (user_tag_id),
	CONSTRAINT user_tag__user_tag_partner_id__unique UNIQUE (user_tag_partner_id, resource_path),
	CONSTRAINT user_tag__user_tag_type__check CHECK ((user_tag_type = ANY (ARRAY['USER_TAG_TYPE_STUDENT'::text, 'USER_TAG_TYPE_STUDENT_DISCOUNT'::text, 'USER_TAG_TYPE_PARENT'::text, 'USER_TAG_TYPE_PARENT_DISCOUNT'::text])))
);

-- bob.tagged_user definition

CREATE TABLE IF NOT EXISTS bob.tagged_user (
	user_id text NOT NULL,
	tag_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL,
	CONSTRAINT pk__tagged_user PRIMARY KEY (user_id, tag_id)
);

-- bob.user_phone_number definition

CREATE TABLE IF NOT EXISTS bob.user_phone_number (
	user_phone_number_id text NOT NULL,
	user_id text NOT NULL,
	phone_number text NOT NULL,
	"type" text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL,
	CONSTRAINT user_phone_number__pk PRIMARY KEY (user_phone_number_id)
);

-- bob.user_address definition

CREATE TABLE IF NOT EXISTS bob.user_address (
	user_address_id text NOT NULL,
	user_id text NOT NULL,
	address_type text NOT NULL,
	postal_code text NULL,
	prefecture_id text NULL,
	city text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NULL,
	first_street text NULL,
	second_street text NULL,
	CONSTRAINT user_address__pk PRIMARY KEY (user_address_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.school_level;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.grade;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.school_level_grade;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.school_course;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.school_info;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.school_history;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.user_tag;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.tagged_user;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.user_phone_number;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.user_address;

CREATE TABLE IF NOT EXISTS bob.school_level_public_info (
    level_id text NOT NULL,
	level_name text NOT NULL,
	school_level_sequence int4 NOT NULL,
	school_level_created_at timestamptz NOT NULL,
	school_level_updated_at timestamptz NOT NULL,
	school_level_deleted_at timestamptz NULL,
	school_level_is_archived bool NOT NULL DEFAULT false,

	grade_id text NOT NULL,
	name text NOT NULL,
	partner_internal_id varchar(50) NOT NULL,
	grade_sequence int4 NOT NULL,
	grade_created_at timestamptz NOT NULL,
	grade_updated_at timestamptz NOT NULL,
	grade_deleted_at timestamptz NULL,
	grade_is_archived bool NOT NULL DEFAULT false,

	school_level_grade_created_at timestamptz NOT NULL,
	school_level_grade_updated_at timestamptz NOT NULL,
	school_level_grade_deleted_at timestamptz NULL,

    CONSTRAINT pk__school_level PRIMARY KEY (level_id)
);

CREATE TABLE IF NOT EXISTS bob.school_course_school_info_public_info (
    school_course_id text NOT NULL,
	school_course_name text NOT NULL,
	school_course_name_phonetic text NULL,
	school_id text NOT NULL,
	school_course_is_archived bool NOT NULL,
	school_course_partner_id text NOT NULL,
	school_course_created_at timestamptz NOT NULL,
	school_course_updated_at timestamptz NOT NULL,
	school_course_deleted_at timestamptz NULL,

	school_name text NOT NULL,
	school_name_phonetic text NULL,
	school_level_id text NOT NULL,
	school_info_is_archived bool NOT NULL,
	school_info_created_at timestamptz NOT NULL,
	school_info_updated_at timestamptz NOT NULL,
	school_info_deleted_at timestamptz NULL,
	school_partner_id text NOT NULL,
	address text NULL,

    CONSTRAINT pk__school_course_school_info PRIMARY KEY (school_id)
);

CREATE TABLE IF NOT EXISTS bob.school_history_public_info (
	student_id text NOT NULL,
	school_id text NOT NULL,
	school_course_id text NULL,
	start_date timestamptz NULL,
	end_date timestamptz NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	is_current bool NULL DEFAULT false,

    CONSTRAINT pk__school_history PRIMARY KEY (student_id, school_id)
);

CREATE TABLE IF NOT EXISTS bob.tagged_user_public_info (
	tag_id text NOT NULL,
	user_id text NOT NULL,
	tagged_user_created_at timestamptz NOT NULL,
	tagged_user_updated_at timestamptz NOT NULL,
	tagged_user_deleted_at timestamptz NULL,
	user_tag_partner_id text NOT NULL,
	name text NOT NULL,
	is_archived bool NOT NULL,
	user_tag_created_at timestamptz NOT NULL,
	user_tag_updated_at timestamptz NOT NULL,
	user_tag_deleted_at timestamptz NULL,
	user_tag_type text NOT NULL,
    
    CONSTRAINT pk__tagged_user PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE IF NOT EXISTS bob.user_phone_number_public_info (
    user_phone_number_id text NOT NULL,
	user_id text NOT NULL,
	phone_number text NOT NULL,
	"type" text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	
	CONSTRAINT pk__user_phone_number PRIMARY KEY (user_phone_number_id)
);

CREATE TABLE IF NOT EXISTS bob.user_address_public_info (
    student_address_id text NOT NULL,
	student_id text NOT NULL,
	address_type text NOT NULL,
	postal_code text NULL,
	prefecture_id text NULL,
	city text NULL,
	user_address_created_at timestamptz NOT NULL,
	user_address_updated_at timestamptz NOT NULL,
	user_address_deleted_at timestamptz NULL,
	resource_path text NULL,
	first_street text NULL,
	second_street text NULL,
	
	CONSTRAINT pk__user_address PRIMARY KEY (student_address_id)
);

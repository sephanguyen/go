ALTER TABLE bob.school_level_public_info 
DROP CONSTRAINT IF EXISTS pk__school_level;
ALTER TABLE bob.school_course_school_info_public_info 
DROP CONSTRAINT IF EXISTS pk__school_course_school_info;
ALTER TABLE bob.permission_public_info
DROP CONSTRAINT IF EXISTS pk__permission;
ALTER TABLE bob.school_history_public_info
DROP CONSTRAINT IF EXISTS pk__school_history;
ALTER TABLE bob.tagged_user_public_info
DROP CONSTRAINT IF EXISTS pk__tagged_user;
ALTER TABLE bob.student_enrollment_status_history_public_info
DROP CONSTRAINT IF EXISTS pk__student_enrollment_status_history;
ALTER TABLE bob.user_group_member_public_info
DROP CONSTRAINT IF EXISTS pk__user_group_member;

CREATE TABLE IF NOT EXISTS bob.school_level (
    level_id text,
	level_name text,
	school_level_sequence int4,
	school_level_created_at timestamptz,
	school_level_updated_at timestamptz,
	school_level_deleted_at timestamptz,
	school_level_is_archived bool,

	grade_id text,
	name text,
	partner_internal_id varchar(50),
	grade_sequence int4,
	grade_created_at timestamptz,
	grade_updated_at timestamptz,
	grade_deleted_at timestamptz,
	grade_is_archived bool,

	school_level_grade_created_at timestamptz,
	school_level_grade_updated_at timestamptz,
	school_level_grade_deleted_at timestamptz,

    CONSTRAINT pk__school_level PRIMARY KEY (level_id, grade_id)
);

CREATE TABLE IF NOT EXISTS bob.school_course_school_info (
    school_course_id text,
	school_course_name text,
	school_course_name_phonetic text,
	school_id text,
	school_course_is_archived bool,
	school_course_partner_id text,
	school_course_created_at timestamptz,
	school_course_updated_at timestamptz,
	school_course_deleted_at timestamptz,

	school_name text,
	school_name_phonetic text,
	school_level_id text,
	school_info_is_archived bool,
	school_info_created_at timestamptz,
	school_info_updated_at timestamptz,
	school_info_deleted_at timestamptz,
	school_partner_id text,
	address text,

    CONSTRAINT pk__school_course_school_info PRIMARY KEY (school_id, school_course_id)
);

CREATE TABLE IF NOT EXISTS bob.permission (
    permission_id text,
	permission_permission_name text,
	permission_created_at timestamptz,
	permission_updated_at timestamptz,
	permission_deleted_at timestamptz,

    permission_role_role_id text,
	permission_role_created_at timestamptz,
	permission_role_updated_at timestamptz,
	permission_role_deleted_at timestamptz,

    user_group_id text,
	user_group_name text,
	granted_permission_role_id text,
	role_name text,
	granted_permission_permission_name text,
	location_id text,

    CONSTRAINT pk__permission PRIMARY KEY (permission_id,user_group_id,granted_permission_role_id,location_id)
);

CREATE TABLE IF NOT EXISTS bob.school_history (
	student_id text,
	school_id text,
	school_course_id text,
	start_date timestamptz,
	end_date timestamptz,
	created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,
	is_current bool DEFAULT false,

    CONSTRAINT pk__school_history PRIMARY KEY (student_id, school_id)
);

CREATE TABLE IF NOT EXISTS bob.tagged_user (
	tag_id text,
	user_id text,
	tagged_user_created_at timestamptz,
	tagged_user_updated_at timestamptz,
	tagged_user_deleted_at timestamptz,
	user_tag_partner_id text,
	name text,
	is_archived bool,
	user_tag_created_at timestamptz,
	user_tag_updated_at timestamptz,
	user_tag_deleted_at timestamptz,
	user_tag_type text,
    
    CONSTRAINT pk__tagged_user PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE IF NOT EXISTS bob.student_enrollment_status_history (
    student_id text,
	location_id text,
	enrollment_status text,
	start_date timestamptz,
	end_date timestamptz,
	"comment" text,
	created_at timestamptz DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz,
	order_id text,
	order_sequence_number int4,

    CONSTRAINT pk__student_enrollment_status_history PRIMARY KEY (student_id,location_id,enrollment_status,start_date)
);

CREATE TABLE IF NOT EXISTS bob.user_group_member (
    user_id text,
	user_group_id text,
	created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);


ALTER PUBLICATION kec_publication SET TABLE 
bob.school_level,    
bob.school_course_school_info,
bob.permission,
bob.school_history,
bob.tagged_user,
bob.student_enrollment_status_history,
bob.user_group_member
;

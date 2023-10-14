CREATE TABLE IF NOT EXISTS public.parents (
    parent_id text,
    parents_created_at timestamptz,
    parents_updated_at timestamptz,
    parents_deleted_at timestamptz,

    user_id text,
    country text,
    name text,
    avatar text,
    phone_number text,
    email text,
    device_token text,
    allow_notification text,
    user_group text,
    users_created_at timestamptz,
    users_updated_at timestamptz,
    users_deleted_at timestamptz,
    given_name text,
    last_login_date timestamptz,
    birthday date,
    gender text,
    first_name text,
    last_name text,
    first_name_phonetic text,
    last_name_phonetic text,
    full_name_phonetic text,
    remarks text,
    is_system bool,
    user_external_id text,
    users_deactivated_at timestamptz,

    CONSTRAINT pk__parents PRIMARY KEY (parent_id)
);

CREATE TABLE IF NOT EXISTS public.staff (
    staff_id text,
    staff_created_at timestamptz,
    staff_updated_at timestamptz,
    staff_deleted_at timestamptz,
    working_status text,
    start_date date,
    end_date date,

    user_id text,
    country text,
    name text,
    avatar text,
    phone_number text,
    email text,
    device_token text,
    allow_notification text,
    user_group text,
    users_created_at timestamptz,
    users_updated_at timestamptz,
    users_deleted_at timestamptz,
    given_name text,
    last_login_date timestamptz,
    birthday date,
    gender text,
    first_name text,
    last_name text,
    first_name_phonetic text,
    last_name_phonetic text,
    full_name_phonetic text,
    remarks text,
    is_system bool,
    user_external_id text,
    users_deactivated_at timestamptz,

    CONSTRAINT pk__staff PRIMARY KEY (staff_id)
);

CREATE TABLE IF NOT EXISTS public.students (
    student_id text,
    students_birthday date,
    students_created_at timestamptz,
    students_updated_at timestamptz,
    students_deleted_at timestamptz,
    school_id int4, 
    student_note text,
    contact_preference text,
    grade_id text,

    user_id text,
    country text,
    name text,
    avatar text,
    phone_number text,
    email text,
    device_token text,
    allow_notification text,
    user_group text,
    users_created_at timestamptz,
    users_updated_at timestamptz,
    users_deleted_at timestamptz,
    given_name text,
    last_login_date timestamptz,
    birthday date,
    gender text,
    first_name text,
    last_name text,
    first_name_phonetic text,
    last_name_phonetic text,
    full_name_phonetic text,
    remarks text,
    is_system bool,
    user_external_id text,
    users_deactivated_at timestamptz,

    CONSTRAINT pk__students PRIMARY KEY (student_id)
);

CREATE TABLE IF NOT EXISTS public.student_parent (
    student_id text,
    parent_id text,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    relationship text,

    CONSTRAINT pk__student_parent PRIMARY KEY (student_id, parent_id)
);

CREATE TABLE IF NOT EXISTS public.role (
    role_id text,
	role_name text,
	role_created_at timestamptz,
	role_updated_at timestamptz,
	role_deleted_at timestamptz,

    granted_role_id text,
	user_group_id text,
	granted_role_created_at timestamptz,
	granted_role_updated_at timestamptz,
	granted_role_deleted_at timestamptz,

	location_id text,
	granted_role_access_path_created_at timestamptz,
	granted_role_access_path_updated_at timestamptz,
	granted_role_access_path_deleted_at timestamptz,

    CONSTRAINT pk__role PRIMARY KEY (role_id, granted_role_id, location_id)
);

CREATE TABLE IF NOT EXISTS public.school_level (
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

CREATE TABLE IF NOT EXISTS public.school_course_school_info (
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

CREATE TABLE IF NOT EXISTS public.permission (
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

CREATE TABLE IF NOT EXISTS public.school_history (
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

CREATE TABLE IF NOT EXISTS public.tagged_user (
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

CREATE TABLE IF NOT EXISTS public.student_enrollment_status_history (
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

CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id text,
	user_group_id text,
	created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);

CREATE TABLE IF NOT EXISTS public.user_group (
    user_group_id text,
	user_group_name text,
	created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,
	resource_path text,
	org_location_id text,
	is_system bool,

    CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);

CREATE TABLE IF NOT EXISTS public.user_phone_number (
    user_phone_number_id text,
	user_id text,
	phone_number text,
	"type" text,
	updated_at timestamptz,
	created_at timestamptz,
	deleted_at timestamptz,
	
	CONSTRAINT pk__user_phone_number PRIMARY KEY (user_phone_number_id)
);

CREATE TABLE IF NOT EXISTS public.user_address (
    student_address_id text,
	student_id text,
	address_type text,
	postal_code text,
	prefecture_id text,
	city text,
	user_address_created_at timestamptz,
	user_address_updated_at timestamptz,
	user_address_deleted_at timestamptz,
	resource_path text,
	first_street text,
	second_street text,
	
	CONSTRAINT pk__user_address PRIMARY KEY (student_address_id)
);


ALTER PUBLICATION kec_publication SET TABLE 
public.user_group,    
public.user_phone_number,
public.user_address,
public.school_level,    
public.school_course_school_info,
public.permission,
public.school_history,
public.tagged_user,
public.student_enrollment_status_history,
public.user_group_member,
public.parents,    
public.staff,
public.students,
public.student_parent,
public.role
;

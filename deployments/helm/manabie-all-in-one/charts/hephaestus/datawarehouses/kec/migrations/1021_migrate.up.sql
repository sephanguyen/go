ALTER TABLE bob.parents_public_info 
DROP CONSTRAINT IF EXISTS pk__parents;
ALTER TABLE bob.staff_public_info 
DROP CONSTRAINT IF EXISTS pk__staff;
ALTER TABLE bob.students_public_info 
DROP CONSTRAINT IF EXISTS pk__students;
ALTER TABLE bob.student_parents_public_info 
DROP CONSTRAINT IF EXISTS pk__student_parents;
ALTER TABLE bob.role_public_info 
DROP CONSTRAINT IF EXISTS pk__role;

CREATE TABLE IF NOT EXISTS bob.parents (
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

CREATE TABLE IF NOT EXISTS bob.staff (
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

CREATE TABLE IF NOT EXISTS bob.students (
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

CREATE TABLE IF NOT EXISTS bob.student_parent (
    student_id text,
    parent_id text,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    relationship text,

    CONSTRAINT pk__student_parent PRIMARY KEY (student_id, parent_id)
);

CREATE TABLE IF NOT EXISTS bob.role (
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

ALTER PUBLICATION kec_publication SET TABLE 
bob.parents,    
bob.staff,
bob.students,
bob.student_parent,
bob.role;

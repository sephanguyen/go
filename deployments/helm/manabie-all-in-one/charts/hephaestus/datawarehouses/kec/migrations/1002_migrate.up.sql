CREATE SCHEMA IF NOT EXISTS bob;
CREATE TABLE IF NOT EXISTS bob.parents_public_info (
    parent_id text NOT NULL,
    parents_created_at timestamptz NOT NULL,
    parents_updated_at timestamptz NOT NULL,
    parents_deleted_at timestamptz,

    country text NOT NULL,
    name text NOT NULL,
    avatar text NULL,
    phone_number text NULL,
    email text NULL,
    device_token text NULL,
    allow_notification text NULL,
    user_group text NOT NULL,
    users_created_at timestamptz NOT NULL,
    users_updated_at timestamptz NOT NULL,
    users_deleted_at timestamptz NULL,
    given_name text NULL,
    last_login_date timestamptz NULL,
    birthday date NULL,
    gender text NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    first_name_phonetic text NULL,
    last_name_phonetic text NULL,
    full_name_phonetic text NULL,
    remarks text NULL,
    is_system bool NULL,
    user_external_id text NULL,

    CONSTRAINT pk__parents PRIMARY KEY (parent_id)
);

CREATE TABLE IF NOT EXISTS bob.staff_public_info (
    staff_id text NOT NULL,
    staff_created_at timestamptz NOT NULL,
    staff_updated_at timestamptz NOT NULL,
    staff_deleted_at timestamptz,
    working_status text,
    start_date date,
    end_date date,

    country text NOT NULL,
    name text NOT NULL,
    avatar text NULL,
    phone_number text NULL,
    email text NULL,
    device_token text NULL,
    allow_notification text NULL,
    user_group text NOT NULL,
    users_created_at timestamptz NOT NULL,
    users_updated_at timestamptz NOT NULL,
    users_deleted_at timestamptz NULL,
    given_name text NULL,
    last_login_date timestamptz NULL,
    birthday date NULL,
    gender text NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    first_name_phonetic text NULL,
    last_name_phonetic text NULL,
    full_name_phonetic text NULL,
    remarks text NULL,
    is_system bool NULL,
    user_external_id text NULL,

    CONSTRAINT pk__staff PRIMARY KEY (staff_id)
);

CREATE TABLE IF NOT EXISTS bob.students_public_info (
    student_id text NOT NULL,
    students_birthday date NULL,
    students_created_at timestamptz NOT NULL,
    students_updated_at timestamptz NOT NULL,
    students_deleted_at timestamptz,
    school_id int4 NULL, 
    student_note text NULL,
    contact_preference text NULL,
    grade_id text NULL,

    country text NOT NULL,
    name text NOT NULL,
    avatar text NULL,
    phone_number text NULL,
    email text NULL,
    device_token text NULL,
    allow_notification text NULL,
    user_group text NOT NULL,
    users_created_at timestamptz NOT NULL,
    users_updated_at timestamptz NOT NULL,
    users_deleted_at timestamptz NULL,
    given_name text NULL,
    last_login_date timestamptz NULL,
    birthday date NULL,
    gender text NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    first_name_phonetic text NULL,
    last_name_phonetic text NULL,
    full_name_phonetic text NULL,
    remarks text NULL,
    is_system bool NULL,
    user_external_id text NULL,

    CONSTRAINT pk__students PRIMARY KEY (student_id)
);

CREATE TABLE IF NOT EXISTS bob.student_parents_public_info (
    student_id text NOT NULL,
    parent_id text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz,
    relationship text NOT NULL,

    CONSTRAINT pk__student_parents PRIMARY KEY (student_id, parent_id)
);

CREATE TABLE IF NOT EXISTS public.users
(
    user_id            TEXT                     NOT NULL
        CONSTRAINT users_pk PRIMARY KEY,
    country            TEXT                     NOT NULL,
    name               TEXT                     NOT NULL,
    avatar             TEXT,
    phone_number       TEXT
        CONSTRAINT users_phone_un UNIQUE,
    email              TEXT
        CONSTRAINT users_email_un UNIQUE,
    device_token       TEXT,
    allow_notification BOOLEAN,
    user_group         TEXT                     NOT NULL
        CONSTRAINT users_user_group_check CHECK (user_group = ANY
                                                 (ARRAY ['USER_GROUP_STUDENT'::TEXT, 'USER_GROUP_COACH'::TEXT, 'USER_GROUP_TUTOR'::TEXT, 'USER_GROUP_STAFF'::TEXT, 'USER_GROUP_ADMIN'::TEXT, 'USER_GROUP_TEACHER'::TEXT, 'USER_GROUP_PARENT'::TEXT, 'USER_GROUP_CONTENT_ADMIN'::TEXT, 'USER_GROUP_CONTENT_STAFF'::TEXT, 'USER_GROUP_SALES_ADMIN'::TEXT, 'USER_GROUP_SALES_STAFF'::TEXT, 'USER_GROUP_CS_ADMIN'::TEXT, 'USER_GROUP_CS_STAFF'::TEXT, 'USER_GROUP_SCHOOL_ADMIN'::TEXT, 'USER_GROUP_SCHOOL_STAFF'::TEXT, 'USER_GROUP_ORGANIZATION_MANAGER'::TEXT])),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    is_tester          BOOLEAN,
    facebook_id        TEXT
        CONSTRAINT users_fb_id_un UNIQUE,
    platform           TEXT
        CONSTRAINT users_platform_check CHECK (platform = ANY
                                               ('{PLATFORM_NONE,PLATFORM_IOS,PLATFORM_ANDROID}'::TEXT[])),
    phone_verified     BOOLEAN,
    email_verified     BOOLEAN,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    given_name         TEXT,
    resource_path      TEXT DEFAULT autofillresourcepath(),
    last_login_date    TIMESTAMP WITH TIME ZONE,
    birthday           DATE,
    gender             TEXT
        CONSTRAINT user_gender_check CHECK (gender = ANY ('{MALE,FEMALE}'::TEXT[]))
);

COMMENT ON COLUMN users.is_tester IS 'to distinguish our staff using app as a student or tester testing app as coach, tutor';


CREATE INDEX IF NOT EXISTS users_name_idx ON users (name);


CREATE INDEX IF NOT EXISTS users_given_name ON users (given_name);


CREATE INDEX IF NOT EXISTS users_resource_path_idx ON users (resource_path);


CREATE INDEX IF NOT EXISTS users__created_at__idx_asc_nulls_last ON users (created_at);


CREATE INDEX IF NOT EXISTS users__created_at__idx_desc_nulls_first ON users (created_at DESC);

CREATE policy rls_users ON "users"
    USING (permission_check(resource_path, 'users'))
    WITH CHECK (permission_check(resource_path, 'users'));

ALTER TABLE "users"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "users"
    FORCE ROW LEVEL SECURITY;

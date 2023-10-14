CREATE TABLE IF NOT EXISTS public.school_level (
    school_level_id TEXT NOT NULL,
    school_level_name TEXT NOT NULL,
    sequence TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT school_level__pk PRIMARY KEY (school_level_id)
);

CREATE POLICY rls_school_level ON "school_level"
USING (permission_check(resource_path, 'school_level'))
WITH CHECK (permission_check(resource_path, 'school_level'));

ALTER TABLE "school_level" ENABLE ROW LEVEL security;
ALTER TABLE "school_level" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.school_course (
    school_course_id TEXT NOT NULL,
    school_course_name TEXT NOT NULL,
    school_course_name_phonetic TEXT,
    school_id TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT school_course__pk PRIMARY KEY (school_course_id),
    CONSTRAINT school_course__school_id__fk FOREIGN KEY (school_id) REFERENCES public.school_info(school_id)
);

CREATE POLICY rls_school_course ON "school_course"
USING (permission_check(resource_path, 'school_course'))
WITH CHECK (permission_check(resource_path, 'school_course'));

ALTER TABLE "school_course" ENABLE ROW LEVEL security;
ALTER TABLE "school_course" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.school_level_grade (
    school_level_id TEXT NOT NULL,
    grade_id TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT school_level_grade__pk PRIMARY KEY (school_level_id, grade_id)
);

CREATE POLICY rls_school_level_grade ON "school_level_grade"
USING (permission_check(resource_path, 'school_level_grade'))
WITH CHECK (permission_check(resource_path, 'school_level_grade'));

ALTER TABLE "school_level_grade" ENABLE ROW LEVEL security;
ALTER TABLE "school_level_grade" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.bank (
    bank_id TEXT NOT NULL,
    bank_code BIGINT NOT NULL,
    bank_name TEXT NOT NULL,
    bank_name_phonetic TEXT,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT bank__pk PRIMARY KEY (bank_id)
);

CREATE POLICY rls_bank ON "bank"
USING (permission_check(resource_path, 'bank'))
WITH CHECK (permission_check(resource_path, 'bank'));

ALTER TABLE "bank" ENABLE ROW LEVEL security;
ALTER TABLE "bank" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.bank_branch (
    bank_branch_id TEXT NOT NULL,
    bank_branch_code BIGINT NOT NULL,
    bank_branch_name TEXT NOT NULL,
    bank_branch_phonetic_name TEXT,
    bank_id TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT bank_branch__pk PRIMARY KEY (bank_branch_id),
    CONSTRAINT bank_branch__bank_id__fk FOREIGN KEY (bank_id) REFERENCES public.bank(bank_id)
);

CREATE POLICY rls_bank_branch ON "bank_branch"
USING (permission_check(resource_path, 'bank_branch'))
WITH CHECK (permission_check(resource_path, 'bank_branch'));

ALTER TABLE "bank_branch" ENABLE ROW LEVEL security;
ALTER TABLE "bank_branch" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.user_tag (
    user_tag_id TEXT NOT NULL,
    user_tag_name TEXT NOT NULL,
    user_tag_type TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT user_tag__pk PRIMARY KEY (user_tag_id)
);

CREATE POLICY rls_user_tag ON "user_tag"
USING (permission_check(resource_path, 'user_tag'))
WITH CHECK (permission_check(resource_path, 'user_tag'));

ALTER TABLE "user_tag" ENABLE ROW LEVEL security;
ALTER TABLE "user_tag" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.timesheet_config (
    timesheet_config_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    config_value TEXT NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT timesheet_config__pk PRIMARY KEY (timesheet_config_id)
);

CREATE POLICY rls_timesheet_config ON "timesheet_config"
USING (permission_check(resource_path, 'timesheet_config'))
WITH CHECK (permission_check(resource_path, 'timesheet_config'));

ALTER TABLE "timesheet_config" ENABLE ROW LEVEL security;
ALTER TABLE "timesheet_config" FORCE ROW LEVEL security;

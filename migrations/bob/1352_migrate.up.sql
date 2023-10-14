CREATE TABLE IF NOT EXISTS public.user_basic_info (
    user_id TEXT NOT NULL,
    "name" TEXT,
    first_name TEXT,
    last_name TEXT,
    full_phonetic_name TEXT,
    first_phonetic_name TEXT,
    last_phonetic_name TEXT,
    current_grade TEXT,
    grade_id TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT fk__user_basic_info__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    CONSTRAINT fk__user_basic_info__grade_id FOREIGN KEY (grade_id) REFERENCES public.grade(grade_id),
    CONSTRAINT pk__user_basic_info PRIMARY KEY (user_id)
);


CREATE POLICY rls_user_basic_info ON "user_basic_info"
USING (permission_check(resource_path, 'user_basic_info')) WITH CHECK (permission_check(resource_path, 'user_basic_info'));
CREATE POLICY rls_user_basic_info_restrictive ON "user_basic_info" AS RESTRICTIVE
USING (permission_check(resource_path, 'user_basic_info'))WITH CHECK (permission_check(resource_path, 'user_basic_info'));

ALTER TABLE "user_basic_info" ENABLE ROW LEVEL security;
ALTER TABLE "user_basic_info" FORCE ROW LEVEL security;

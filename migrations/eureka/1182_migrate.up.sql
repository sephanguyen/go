CREATE TABLE IF NOT EXISTS public.grade_book_setting (
    setting text,
    updated_by text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
	resource_path text NULL DEFAULT autofillresourcepath(),
    CONSTRAINT grade_book_setting_pk PRIMARY KEY (setting),
    CONSTRAINT grade_book_setting_check CHECK ((setting = ANY (ARRAY['LATEST_SCORE'::text, 'GRADE_TO_PASS_SCORE'::text])))
);

CREATE POLICY rls_grade_book_setting ON "grade_book_setting" using (
    permission_check(resource_path, 'grade_book_setting')
) with check (
    permission_check(resource_path, 'grade_book_setting')
);

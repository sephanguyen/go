-- public.lesson_reports definition
CREATE TABLE public.lesson_reports (
	lesson_report_id text NOT NULL,
	report_submitting_status text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	form_config_id text NULL,
	lesson_id text NULL,
	CONSTRAINT lesson_reports_pk PRIMARY KEY (lesson_report_id)
);


CREATE POLICY rls_lesson_reports ON "lesson_reports" USING (permission_check(resource_path, 'lesson_reports')) WITH CHECK (permission_check(resource_path, 'lesson_reports'));
CREATE POLICY rls_lesson_reports_restrictive ON "lesson_reports" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_reports')) with check (permission_check(resource_path, 'lesson_reports'));

ALTER TABLE "lesson_reports" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_reports" FORCE ROW LEVEL security;


-- public.lesson_report_details definition
CREATE TABLE public.lesson_report_details (
	lesson_report_id text NOT NULL,
	student_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	lesson_report_detail_id text NOT NULL,
	CONSTRAINT lesson_report_details_pk PRIMARY KEY (lesson_report_detail_id),
	CONSTRAINT unique__lesson_report_id__student_id UNIQUE (lesson_report_id, student_id)
);

CREATE POLICY rls_lesson_report_details ON "lesson_report_details" USING (permission_check(resource_path, 'lesson_report_details')) WITH CHECK (permission_check(resource_path, 'lesson_report_details'));
CREATE POLICY rls_lesson_report_details_restrictive ON "lesson_report_details" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_report_details')) with check (permission_check(resource_path, 'lesson_report_details'));

ALTER TABLE "lesson_report_details" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_report_details" FORCE ROW LEVEL security;
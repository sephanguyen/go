CREATE TABLE IF NOT EXISTS bob.lesson_reports (
	lesson_report_id text NOT NULL,
	report_submitting_status text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	form_config_id text NULL,
	lesson_id text NULL,
	CONSTRAINT lesson_reports_pk PRIMARY KEY (lesson_report_id)
);

CREATE TABLE IF NOT EXISTS bob.lesson_report_details (
	lesson_report_id text NOT NULL,
	student_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	lesson_report_detail_id text NOT NULL,
	report_versions INTEGER DEFAULT 0,
	CONSTRAINT lesson_report_details_pk PRIMARY KEY (lesson_report_detail_id),
	CONSTRAINT unique__lesson_report_id__student_id UNIQUE (lesson_report_id, student_id)
);

CREATE TABLE IF NOT EXISTS bob.classroom (
    classroom_id TEXT NOT NULL,
    name TEXT NOT NULL,
    location_id TEXT NOT NULL,
    remarks TEXT NULL,
    is_archived boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    resource_path text NOT NULL,

    CONSTRAINT classroom_pk PRIMARY KEY (classroom_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.lesson_reports;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.lesson_report_details;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.classroom;

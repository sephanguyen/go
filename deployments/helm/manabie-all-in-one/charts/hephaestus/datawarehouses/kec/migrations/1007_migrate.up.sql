CREATE TABLE IF NOT EXISTS bob.classroom_public_info (
    classroom_id TEXT NOT NULL,
    name TEXT NOT NULL,
    location_id TEXT NOT NULL,
    remarks TEXT NULL,
    is_archived boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
    CONSTRAINT classroom_public_info_pk PRIMARY KEY (classroom_id)
);

CREATE TABLE IF NOT EXISTS bob.lesson_reports_public_info (
    lesson_report_detail_id text NOT NULL,
	lesson_report_id text NOT NULL,
	student_id text NOT NULL,
	lesson_report_details_created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	lesson_report_details_updated_at timestamptz NULL,
	lesson_report_details_deleted_at timestamptz NULL,
    report_submitting_status text NOT NULL,
    lesson_reports_created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	lesson_reports_updated_at timestamptz NULL,
	lesson_reports_deleted_at timestamptz NULL,
	form_config_id text NULL,
	lesson_id text NULL,
	CONSTRAINT lesson_report_public_info_pk PRIMARY KEY (lesson_report_detail_id),
	CONSTRAINT unique__lesson_report_id__student_id_public_info UNIQUE (lesson_report_id, student_id)
);

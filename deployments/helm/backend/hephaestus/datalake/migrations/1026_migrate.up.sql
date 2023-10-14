CREATE SCHEMA IF NOT EXISTS lessonmgmt;

CREATE TABLE IF NOT EXISTS lessonmgmt.lessons_teachers (
    lesson_id text NOT NULL,
    teacher_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL,
	teacher_name text NULL,
    CONSTRAINT lessons_teachers_pk PRIMARY KEY (lesson_id, teacher_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.lessons_courses (
    lesson_id text NOT NULL,
    course_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL,
    CONSTRAINT lessons_courses_pk PRIMARY KEY (lesson_id, course_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.reallocation (
    student_id TEXT NULL,
    original_lesson_id TEXT NOT NULL,
    new_lesson_id TEXT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    course_id TEXT NOT NULL,
    resource_path TEXT NOT NULL,
    CONSTRAINT pk__reallocation PRIMARY KEY (original_lesson_id, student_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.lesson_reports (
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

CREATE TABLE IF NOT EXISTS lessonmgmt.lesson_report_details (
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

CREATE TABLE IF NOT EXISTS lessonmgmt.classroom (
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

CREATE TABLE IF NOT EXISTS lessonmgmt.partner_form_configs (
    form_config_id TEXT NOT NULL,
    partner_id INTEGER NOT NULL,
    feature_name TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    form_config_data JSONB NULL,
    resource_path TEXT NOT NULL,
    CONSTRAINT pk__partner_form_configs PRIMARY KEY (form_config_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.partner_dynamic_form_field_values (
	dynamic_form_field_value_id text NOT NULL,
	field_id text NOT NULL,
	lesson_report_detail_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	value_type text NULL,
	string_value text NULL,
	bool_value bool NULL,
	string_array_value _text NULL,
	int_array_value _int4 NULL,
	string_set_value _text NULL,
	int_set_value _int4 NULL,
	field_render_guide jsonb NULL,
	resource_path text NOT NULL,
	int_value int4 NULL,
	CONSTRAINT partner_dynamic_form_field_values_pk PRIMARY KEY (dynamic_form_field_value_id),
	CONSTRAINT unique__lesson_report_detail_id__field_id UNIQUE (lesson_report_detail_id, field_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.lessons (
	lesson_id text NOT NULL,
	teacher_id text NULL,
	course_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	end_at timestamptz NULL,
	control_settings jsonb NULL,
	lesson_group_id text NULL,
	room_id text NULL,
	lesson_type text NULL,
	status text NULL,
	stream_learner_counter int4 NOT NULL,
	learner_ids _text NOT NULL,
	name text NULL,
	start_time timestamptz NULL,
	end_time timestamptz NULL,
	resource_path text NOT NULL,
	room_state jsonb NULL,
	teaching_model text NULL,
	class_id text NULL,
	center_id text NULL,
	teaching_method text NULL,
	teaching_medium text NULL,
	scheduling_status text NULL,
	is_locked bool NOT NULL,
	scheduler_id text NULL,
	zoom_link TEXT,
	zoom_owner_id TEXT,
	zoom_id TEXT,
	zoom_occurrence_id TEXT,
	CONSTRAINT lessons_pk PRIMARY KEY (lesson_id)
);

CREATE TABLE IF NOT EXISTS lessonmgmt.lesson_members (
	lesson_id text NOT NULL,
	user_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	attendance_status text NULL,
	attendance_remark text NULL,
	course_id text NULL,
	attendance_notice text NULL,
	attendance_reason text NULL,
	attendance_note text NULL,
	user_first_name text NULL,
	user_last_name text NULL,
	CONSTRAINT pk__lesson_members PRIMARY KEY (lesson_id, user_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
lessonmgmt.lessons_teachers,
lessonmgmt.lessons_courses,
lessonmgmt.reallocation,
lessonmgmt.lesson_reports,
lessonmgmt.lesson_report_details,
lessonmgmt.partner_form_configs,
lessonmgmt.partner_dynamic_form_field_values,
lessonmgmt.classroom,
lessonmgmt.lessons,
lessonmgmt.lesson_members;

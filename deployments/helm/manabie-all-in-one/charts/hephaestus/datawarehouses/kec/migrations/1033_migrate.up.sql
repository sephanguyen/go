CREATE TABLE IF NOT EXISTS public.scheduler (
    scheduler_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    freq frequency,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT pk__scheduler PRIMARY KEY (scheduler_id)
);

CREATE TABLE IF NOT EXISTS public.lessons_teachers (
    lesson_id text NOT NULL,
    staff_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT lesson_teachers_pk PRIMARY KEY (lesson_id, staff_id)
);

CREATE TABLE IF NOT EXISTS public.lessons_courses (
    lesson_id text NOT NULL,
    course_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT lesson_courses_pk PRIMARY KEY (lesson_id, course_id)
);

CREATE TABLE IF NOT EXISTS public.reallocation (
    student_id TEXT NULL,
    original_lesson_id TEXT NOT NULL,
    new_lesson_id TEXT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    course_id TEXT NOT NULL,
    CONSTRAINT pk__reallocation PRIMARY KEY (original_lesson_id, student_id)
);

CREATE TABLE IF NOT EXISTS public.classroom (
    classroom_id TEXT NOT NULL,
    name TEXT NOT NULL,
    location_id TEXT NOT NULL,
    remarks TEXT NULL,
    is_archived boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
    CONSTRAINT classroom_pk PRIMARY KEY (classroom_id)
);

CREATE TABLE IF NOT EXISTS public.lesson_reports (
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
	CONSTRAINT lesson_report_pk PRIMARY KEY (lesson_report_detail_id),
	CONSTRAINT unique__lesson_report_id__student_id UNIQUE (lesson_report_id, student_id)
);

CREATE TABLE IF NOT EXISTS public.partner_form_configs (
    form_config_id TEXT NOT NULL,
    partner_id INTEGER NOT NULL,
    feature_name TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    form_config_data JSONB NULL,
    CONSTRAINT pk__partner_form_configs PRIMARY KEY (form_config_id)
);

CREATE TABLE IF NOT EXISTS public.partner_dynamic_form_field_values (
	dynamic_form_field_value_id text NOT NULL,
	field_id text NOT NULL,
	lesson_report_detail_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	value_type text NULL,
	string_value text NULL,
	int_value int4 NULL,
	CONSTRAINT partner_dynamic_form_field_values_pk PRIMARY KEY (dynamic_form_field_value_id),
	CONSTRAINT unique__lesson_report_detail_id__field_id UNIQUE (lesson_report_detail_id, field_id)
);

CREATE TABLE IF NOT EXISTS public.day_info (
	"date" date NOT NULL,
	location_id text NOT NULL,
	day_type_id text NULL,
	opening_time text NULL,
	status day_info_status NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	time_zone text NULL DEFAULT current_setting('TIMEZONE'::text),
	CONSTRAINT day_info_pk PRIMARY KEY (location_id, date)
);

CREATE TABLE IF NOT EXISTS public.day_type (
	day_type_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    display_name text NULL,
    is_archived boolean NOT NULL DEFAULT false,
	CONSTRAINT day_type_pk PRIMARY KEY (day_type_id)
);

CREATE TABLE IF NOT EXISTS public.lesson_members (
    lesson_id text NOT NULL,
    student_id text NOT NULL,
    course_id text,
    lesson_members_updated_at timestamptz NOT NULL,
    lesson_members_created_at timestamptz NOT NULL,
    lesson_members_deleted_at timestamptz,
    attendance_status text,
    attendance_remark text,
    attendance_notice text,
    attendance_reason text,
    attendance_note text,
    lessons_created_at timestamptz NOT NULL,
    lessons_updated_at timestamptz NOT NULL,
    lessons_deleted_at timestamptz,
    end_at timestamptz,
    control_settings jsonb null,
    lesson_group_id text,
    room_id text,
    lesson_type text,
    status text,
    stream_learner_counter int4 NOT NULL,
    learner_ids _text NOT NULL DEFAULT '{}'::text[],
    name text,
    start_time timestamptz NULL,
    end_time timestamptz NULL,
    room_state jsonb NULL,
    class_id text,
    center_id text,
    teaching_method text,
    teaching_medium text,
    scheduling_status text,
    is_locked bool NOT NULL,
    scheduler_id text,

    CONSTRAINT pk__lesson_members PRIMARY KEY (lesson_id,student_id)
);

ALTER PUBLICATION kec_publication ADD TABLE
public.scheduler,
public.lessons_teachers,
public.lessons_courses,
public.reallocation,
public.classroom,
public.lesson_reports,
public.partner_form_configs,
public.partner_dynamic_form_field_values,
public.day_info,
public.day_type,
public.lesson_members;

DROP TABLE IF EXISTS bob.scheduler_public_info;
DROP TABLE IF EXISTS bob.lessons_teachers_public_info;
DROP TABLE IF EXISTS bob.lessons_courses_public_info;
DROP TABLE IF EXISTS bob.reallocation_public_info;
DROP TABLE IF EXISTS bob.classroom_public_info;
DROP TABLE IF EXISTS bob.lesson_reports_public_info;
DROP TABLE IF EXISTS bob.partner_form_configs_public_info;
DROP TABLE IF EXISTS bob.partner_dynamic_form_field_values_public_info;
DROP TABLE IF EXISTS bob.day_info_public_info;
DROP TABLE IF EXISTS bob.day_type_public_info;
DROP TABLE IF EXISTS bob.lesson_members_public_info;

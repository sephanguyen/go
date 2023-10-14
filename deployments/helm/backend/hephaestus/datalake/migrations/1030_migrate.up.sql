CREATE SCHEMA IF NOT EXISTS calendar;

CREATE TABLE IF NOT EXISTS calendar.day_type (
	day_type_id text NOT NULL,
	resource_path TEXT,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    display_name text NULL,
    is_archived boolean NOT NULL DEFAULT false,
	CONSTRAINT day_type_pk PRIMARY KEY (day_type_id,resource_path)
);

CREATE TABLE IF NOT EXISTS calendar.scheduler (
    scheduler_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    freq frequency,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT NOT NULL,
    CONSTRAINT pk__scheduler PRIMARY KEY (scheduler_id)
);

CREATE TABLE IF NOT EXISTS calendar.day_info (
	"date" date NOT NULL,
	location_id text NOT NULL,
	day_type_id text NULL,
	opening_time text NULL,
	status day_info_status NULL,
	resource_path text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	time_zone text NULL DEFAULT current_setting('TIMEZONE'::text),
	CONSTRAINT day_info_pk PRIMARY KEY (location_id, date)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE calendar.scheduler;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE calendar.day_type;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE calendar.day_info;

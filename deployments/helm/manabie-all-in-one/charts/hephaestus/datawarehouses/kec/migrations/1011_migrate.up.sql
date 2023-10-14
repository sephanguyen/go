CREATE SCHEMA IF NOT EXISTS bob;

CREATE TYPE day_info_status AS enum  ('none', 'draft', 'published');

CREATE TABLE IF NOT EXISTS bob.day_info_public_info (
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

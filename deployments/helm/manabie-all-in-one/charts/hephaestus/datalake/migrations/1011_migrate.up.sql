CREATE SCHEMA IF NOT EXISTS bob;

CREATE TYPE day_info_status AS enum  ('none', 'draft', 'published');

CREATE TABLE IF NOT EXISTS bob.day_info (
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

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.day_info;
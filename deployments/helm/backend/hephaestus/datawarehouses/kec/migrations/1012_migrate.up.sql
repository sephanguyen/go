CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.day_type_public_info (
	day_type_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    display_name text NULL,
    is_archived boolean NOT NULL DEFAULT false,
	CONSTRAINT day_type_pk PRIMARY KEY (day_type_id)
);

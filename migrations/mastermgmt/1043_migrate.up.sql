CREATE SCHEMA IF NOT EXISTS manabie;

CREATE TABLE manabie.user_basic_info (
	user_id text NOT NULL,
	"name" text NULL,
	grade_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT pk__user_basic_info PRIMARY KEY (user_id)
);

CREATE SCHEMA IF NOT EXISTS bob;

CREATE TYPE frequency AS ENUM ('once', 'weekly');

CREATE TABLE IF NOT EXISTS bob.scheduler_public_info (
    scheduler_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    freq frequency,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    test TEXT,
    CONSTRAINT pk__scheduler PRIMARY KEY (scheduler_id)
);

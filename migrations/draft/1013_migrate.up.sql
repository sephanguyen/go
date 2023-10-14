CREATE TABLE IF NOT EXISTS "github_raw_data" (
    "id" SERIAL PRIMARY KEY,
    "event_name" text,
    "data" JSONB
);

CREATE TABLE IF NOT EXISTS "github_pr_statistic" (
    "id" SERIAL PRIMARY KEY,
    "branch_name" text NOT NULL,
    "number" integer NOT NULL,
    "create_at" timestamp with time zone NOT NULL,
    "close_at" timestamp with time zone NOT NULL,
    "number_comments" integer DEFAULT 0 NOT NULL,
    "time_to_first_comment" float DEFAULT 0 NOT NULL,
    "total_time_consuming" float DEFAULT 0 NOT NULL,
    "is_merged" boolean DEFAULT false NOT NULL
);

CREATE TABLE IF NOT EXISTS "history" (
    "id" SERIAL PRIMARY KEY,
    "branch_name" text,
    "repository" text,
    "coverage" real,
    "time" timestamp with time zone,
    "status" text
);


CREATE TABLE IF NOT EXISTS "target_coverage" (
    "id" SERIAL PRIMARY KEY,
    "branch_name" text,
    "coverage" real,
    "repository" text unique,
    "update_at" timestamp with time zone,
    "secret_key" text
);
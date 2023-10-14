CREATE TABLE IF NOT EXISTS "locations" (
    "location_id" TEXT NOT NULL PRIMARY KEY,
    "name" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" TEXT
);

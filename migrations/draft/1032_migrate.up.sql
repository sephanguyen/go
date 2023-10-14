CREATE TABLE IF NOT EXISTS "mfe_services_versions" (
    "id" SERIAL PRIMARY KEY,
    "type" text,
    "version" text NOT NULL,
    "squad_name" text NOT NULL,
    "service_name" text NOT NULL,
    "environment" text NOT NULL,
    "organization" text NOT NULL,
    "link" text,
    "created_at" timestamp with time zone,
    "deployed_at" timestamp with time zone,
    "rollback_at" timestamp with time zone
);


CREATE TABLE IF NOT EXISTS "mfe_import_map_versions" (
    "id" SERIAL PRIMARY KEY,
    "environment" text NOT NULL,
    "organization" text NOT NULL,
    "deployed_at" timestamp with time zone,
    "import_map" JSON NOT NULL,
    "type" text
);
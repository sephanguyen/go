
CREATE TABLE IF NOT EXISTS "e2e_instances" (
    "instance_id" TEXT NOT NULL PRIMARY KEY,
    "metadata" JSONB DEFAULT '{}'::jsonb,
    "total_worker" INTEGER DEFAULT 1,
    "duration" INTEGER DEFAULT 0,
    "status" TEXT,
    "name" TEXT,
    "status_statistics" JSONB DEFAULT '{}',
    "flavor" JSONB DEFAULT '{}',
    "tags" TEXT[],
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "started_at" timestamp with time zone,
    "ended_at" timestamp with time zone,
    "deleted_at" timestamp with time zone
);


CREATE TABLE IF NOT EXISTS "e2e_features" (
    "feature_id" TEXT NOT NULL PRIMARY KEY,
    "instance_id" TEXT NOT NULL,
    "worker_id" INTEGER DEFAULT 0,
    "duration" INTEGER DEFAULT 0,
    "status" TEXT,
    "uri" TEXT,
    "data" TEXT,
    "keyword" TEXT,
    "name" TEXT,
    "media_type" TEXT,
    "rules" TEXT[],
    "description" TEXT,
    "scenarios" JSONB DEFAULT '{}'::jsonb,
    "background" JSONB DEFAULT '{}'::jsonb,
    "elements" TEXT[],
    "tags" TEXT[],

    "children" JSONB DEFAULT '{}'::jsonb,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "started_at" timestamp with time zone,
    "ended_at" timestamp with time zone
);

CREATE TABLE IF NOT EXISTS "e2e_scenarios" (
    "scenario_id" TEXT NOT NULL PRIMARY KEY,
    "feature_id" TEXT NOT NULL,
   
    "tags"  TEXT[],
    "keyword" TEXT,
    "name" TEXT,
    "description" TEXT,
    "steps" JSONB DEFAULT '{}'::jsonb,
    "status" TEXT,

    "pickle" JSONB DEFAULT '{}'::jsonb,
    "test_case" JSONB DEFAULT '{}'::jsonb,

    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "started_at" timestamp with time zone,
    "ended_at" timestamp with time zone
);

CREATE TABLE IF NOT EXISTS "e2e_steps" (
    "step_id" TEXT NOT NULL PRIMARY KEY,
    "scenario_id" TEXT NOT NULL,
    "duration" numeric DEFAULT 0,
    "index" smallint DEFAULT 0,
    "status" TEXT,
    "keyword" TEXT,
    "uri" TEXT,
    "name" TEXT,
    "type" TEXT,
    "message" TEXT,
    "is_hook" BOOLEAN DEFAULT false,
    "will_be_retried" BOOLEAN DEFAULT false,
    "embeddings" JSONB DEFAULT '[]'::jsonb,

    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "started_at" timestamp with time zone,
    "ended_at" timestamp with time zone,
    "deleted_at" timestamp with time zone
);


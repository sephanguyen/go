CREATE TABLE IF NOT EXISTS "orgs" (
    "org_id" text NOT NULL,
    "name" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    CONSTRAINT orgs_pk PRIMARY KEY (org_id),
    CONSTRAINT orgs_un UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS "brands" (
    "brand_id" text NOT NULL,
    "name" text NOT NULL,
    "org_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "owner" text,
    CONSTRAINT brands_pk PRIMARY KEY (brand_id),
    CONSTRAINT brands_orgs_fk FOREIGN KEY (org_id) REFERENCES public.orgs(org_id),
    CONSTRAINT brands_un UNIQUE (name, org_id)
);

CREATE TABLE IF NOT EXISTS "centers" (
    "center_id" text NOT NULL,
    "name" text NOT NULL,
    "brand_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "owner" text,
    CONSTRAINT centers_pk PRIMARY KEY (center_id),
    CONSTRAINT centers_brands_fk FOREIGN KEY (brand_id) REFERENCES public.brands(brand_id),
    CONSTRAINT centers_un UNIQUE (name, brand_id)
);

CREATE TABLE IF NOT EXISTS "scheduler_patterns" (
    "scheduler_pattern_id" text NOT NULL,
    "scheduler_pattern_parent_id" text,
    "scheduler_type" text NOT NULL, -- ["opening_time" | "event"]
    "time_zone" text NOT NULL,
    "start_time" timestamp with time zone NOT NULL,
    "end_time" timestamp with time zone NOT NULL,
    "all_day" boolean DEFAULT false,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "meta_data" jsonb,
    "repeat_option_data" jsonb,
    "latest_released_at" timestamp with time zone,
    "brand_id" text,
    "center_id" text,
    "owner" text,
    CONSTRAINT scheduler_patterns_pk PRIMARY KEY (scheduler_pattern_id),
    CONSTRAINT scheduler_patterns_scheduler_patterns_fk FOREIGN KEY (scheduler_pattern_parent_id) REFERENCES public.scheduler_patterns(scheduler_pattern_id),
    CONSTRAINT scheduler_patterns_brands_fk FOREIGN KEY (brand_id) REFERENCES public.brands(brand_id),
    CONSTRAINT scheduler_patterns_centers_fk FOREIGN KEY (center_id) REFERENCES public.centers(center_id),
    CONSTRAINT chk_parent check ((brand_id is not null and center_id is null) or (brand_id is null and center_id is not null))
);

CREATE TABLE IF NOT EXISTS "scheduler_items" (
    "scheduler_item_id" text,
    "scheduler_pattern_id" text NOT NULL,
    "start_time" timestamp with time zone NOT NULL,
    "end_time" timestamp with time zone NOT NULL,
    "all_day" boolean DEFAULT false,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "meta_data" jsonb,
    "brand_id" text,
    "center_id" text,
    "owner" text,
    CONSTRAINT scheduler_items_pk PRIMARY KEY (scheduler_item_id),
    CONSTRAINT scheduler_items_scheduler_patterns_fk FOREIGN KEY (scheduler_pattern_id) REFERENCES public.scheduler_patterns(scheduler_pattern_id),
    CONSTRAINT scheduler_items_brands_fk FOREIGN KEY (brand_id) REFERENCES public.brands(brand_id),
    CONSTRAINT scheduler_items_centers_fk FOREIGN KEY (center_id) REFERENCES public.centers(center_id)
);

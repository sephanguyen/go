CREATE TABLE IF NOT EXISTS "lo_study_plan_items" (
    "lo_id" text,
    "study_plan_item_id" text,
    "created_at" timestamp with time zone NOT NULL,
    "updated_at" timestamp with time zone NOT NULL,
    "deleted_at" timestamp with time zone,
    CONSTRAINT lo_study_plan_items_pk PRIMARY KEY (study_plan_item_id, lo_id)
);

ALTER TABLE "lo_study_plan_items" ADD FOREIGN KEY ("study_plan_item_id") REFERENCES "study_plan_items" ("study_plan_item_id");

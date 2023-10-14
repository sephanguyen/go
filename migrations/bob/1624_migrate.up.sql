ALTER TABLE "courses"
    ADD COLUMN is_adaptive boolean default false,
    ADD COLUMN vendor_id text NULL;

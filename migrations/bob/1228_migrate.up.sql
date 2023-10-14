-- enum date_info_status rename
ALTER TYPE day_info_status RENAME TO date_info_status;

-- public.date_info rename
ALTER TABLE IF EXISTS day_info RENAME TO date_info;
-- alter status column definition: check if need to or not

-- rename date_info table constraints
ALTER TABLE IF EXISTS date_info RENAME CONSTRAINT day_info_pk TO date_info_pk;
ALTER TABLE IF EXISTS date_info RENAME CONSTRAINT day_info_fk TO date_info_fk;

-- rename policy and re-enforce 
DROP POLICY rls_day_info ON date_info;

CREATE POLICY rls_date_info ON "date_info" using (permission_check(resource_path, 'date_info')) with check (permission_check(resource_path, 'date_info'));

ALTER TABLE "date_info" ENABLE ROW LEVEL security;
ALTER TABLE "date_info" FORCE ROW LEVEL security;

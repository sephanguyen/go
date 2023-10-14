-- -- user_basic_info
-- ALTER TABLE manabie.user_basic_info ADD COLUMN IF NOT EXISTS user_role TEXT DEFAULT NULL;

-- DROP INDEX IF EXISTS user_basic_info__user_role__idx;
-- CREATE INDEX user_basic_info__user_role__idx ON manabie.user_basic_info (user_role);

select 1;

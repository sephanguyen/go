ALTER TABLE IF EXISTS address_street DROP CONSTRAINT address_street__user_address_id__fk;
ALTER TABLE IF EXISTS address_street DROP CONSTRAINT address_street__pk;

DROP TABLE IF EXISTS address_street;

ALTER TABLE IF EXISTS user_address ADD first_street TEXT;
ALTER TABLE IF EXISTS user_address ADD second_street TEXT;
ALTER TABLE IF EXISTS user_address RENAME type TO address_type;
--Drop existing constraint
--Add new constraint
ALTER TABLE IF EXISTS usr_email DROP CONSTRAINT IF EXISTS usr_email__pkey;
ALTER TABLE IF EXISTS usr_email ADD CONSTRAINT usr_email__pkey UNIQUE (email, resource_path);


--Drop existing constraint
ALTER TABLE IF EXISTS apple_users DROP CONSTRAINT IF EXISTS pk__apple_users;
--Add new constraint
ALTER TABLE IF EXISTS apple_users DROP CONSTRAINT IF EXISTS apple_usr__pk;
ALTER TABLE IF EXISTS apple_users ADD CONSTRAINT apple_usr__pk PRIMARY KEY (apple_user_id, resource_path);


--Drop existing constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_email_un;
--Add new constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users__email__key;
ALTER TABLE IF EXISTS users ADD CONSTRAINT users__email__key UNIQUE (email, resource_path);

--Drop existing constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_phone_un;
--Add new constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users__phone_number__key;
ALTER TABLE IF EXISTS users ADD CONSTRAINT users__phone_number__key UNIQUE (phone_number, resource_path);

--Drop existing constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_fb_id_un;
--Add new constraint
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users__facebook_id__key;
ALTER TABLE IF EXISTS users ADD CONSTRAINT users__facebook_id__key UNIQUE (facebook_id, resource_path);

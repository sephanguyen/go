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
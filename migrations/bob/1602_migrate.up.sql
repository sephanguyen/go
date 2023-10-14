ALTER TABLE users ADD COLUMN IF NOT EXISTS login_email text;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users__login_email__unique;
ALTER TABLE users ADD CONSTRAINT users__login_email__unique UNIQUE (login_email, resource_path);

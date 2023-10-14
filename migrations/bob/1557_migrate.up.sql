ALTER TABLE users ADD COLUMN IF NOT EXISTS username text;

DROP INDEX IF EXISTS users__username__key;
CREATE UNIQUE INDEX users__username__key ON users (username, resource_path);

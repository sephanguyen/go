ALTER TABLE bob.users ADD COLUMN IF NOT EXISTS deactivated_at timestamptz NULL;

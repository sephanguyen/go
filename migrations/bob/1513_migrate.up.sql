ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS encrypted_user_id_by_password TEXT

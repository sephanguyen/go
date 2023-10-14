ALTER TABLE IF EXISTS bob.staff_public_info ADD COLUMN IF NOT EXISTS users_deactivated_at timestamptz;
ALTER TABLE IF EXISTS bob.students_public_info ADD COLUMN IF NOT EXISTS users_deactivated_at timestamptz;
ALTER TABLE IF EXISTS bob.parents_public_info ADD COLUMN IF NOT EXISTS users_deactivated_at timestamptz;
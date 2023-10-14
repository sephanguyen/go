--- Migrate null value ---
UPDATE user_group
SET is_system = FALSE
WHERE is_system IS NULL;

UPDATE role
SET is_system = TRUE
WHERE is_system IS NULL;

--- Drop default value ---
ALTER TABLE IF EXISTS public.role
    ALTER COLUMN is_system DROP DEFAULT,
    ALTER COLUMN is_system SET NOT NULL;

ALTER TABLE IF EXISTS public.user_group
    ALTER COLUMN is_system DROP DEFAULT,
    ALTER COLUMN is_system SET NOT NULL;

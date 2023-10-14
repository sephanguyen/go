ALTER TABLE questionnaires DROP COLUMN IF EXISTS end_date;
ALTER TABLE questionnaires ADD COLUMN IF NOT EXISTS expiration_date timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now());

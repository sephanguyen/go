ALTER TABLE ONLY student_qr ADD COLUMN IF NOT EXISTS version text;
ALTER TABLE ONLY student_qr	ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc');

ALTER TABLE ONLY student_qr ALTER COLUMN created_at SET DEFAULT (now() at time zone 'utc');
ALTER TABLE ONLY student_qr	ADD CONSTRAINT version_check CHECK ((version = ANY (ARRAY[''::text, 'v2'::text])));
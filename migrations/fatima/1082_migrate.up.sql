ALTER TABLE ONLY public.grade
DROP CONSTRAINT IF EXISTS grade_pk;

ALTER TABLE ONLY public.grade
DROP COLUMN IF EXISTS id;

ALTER TABLE ONLY public.grade
ADD COLUMN IF NOT EXISTS grade_id TEXT NOT NULL;

ALTER TABLE ONLY public.grade
ADD COLUMN IF NOT EXISTS partner_internal_id VARCHAR(50) NOT NULL;

ALTER TABLE ONLY public.grade
ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone NULL;

ALTER TABLE ONLY public.grade
ADD CONSTRAINT grade_pk PRIMARY KEY (grade_id);

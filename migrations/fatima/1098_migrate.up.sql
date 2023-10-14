ALTER TABLE public.students ADD COLUMN grade_id text;

ALTER TABLE product_grade ALTER COLUMN grade_id TYPE text;

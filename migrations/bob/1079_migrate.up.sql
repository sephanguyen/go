DO $$
BEGIN
  IF EXISTS(SELECT *
    FROM information_schema.columns
    WHERE table_name='students' and column_name='status')
  THEN
      ALTER TABLE ONLY public.students RENAME COLUMN status TO enrollment_status;
  END IF;
END $$;

ALTER TABLE ONLY public.students ALTER COLUMN enrollment_status SET DEFAULT 'STUDENT_ENROLLMENT_STATUS_ENROLLED';
ALTER TABLE ONLY public.students DROP CONSTRAINT IF EXISTS students_status_check;
UPDATE public.students SET enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED' WHERE enrollment_status = 'STUDENT_STATUS_ENROLLED';
ALTER TABLE ONLY public.students DROP CONSTRAINT IF EXISTS students_enrollment_status_check;
ALTER TABLE ONLY public.students ADD CONSTRAINT students_enrollment_status_check CHECK ((enrollment_status = ANY ('{STUDENT_ENROLLMENT_STATUS_POTENTIAL, STUDENT_ENROLLMENT_STATUS_ENROLLED, STUDENT_ENROLLMENT_STATUS_QUIT}'::text[])));

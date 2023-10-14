ALTER TABLE IF EXISTS public.bank_branch 
  DROP CONSTRAINT IF EXISTS bank_branch__bank_branch_code__unique,
  ADD CONSTRAINT bank_branch__bank_branch_code__unique UNIQUE (bank_branch_code, bank_id, resource_path);

ALTER TABLE IF EXISTS public.school_history 
  DROP CONSTRAINT IF EXISTS school_history__school_course_id__fk,
  ADD CONSTRAINT school_history__school_course_id__fk FOREIGN KEY (school_course_id) REFERENCES public.school_course(school_course_id);

ALTER TABLE public.student_submissions ADD student_submission_grade_id text NULL;
ALTER TABLE public.student_submissions ADD CONSTRAINT student_submissions_grades_fk FOREIGN KEY (student_submission_grade_id) REFERENCES public.student_submission_grades(student_submission_grade_id);

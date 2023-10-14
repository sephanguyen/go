CREATE INDEX IF NOT EXISTS course_students_access_paths_course_id_idx ON public.course_students_access_paths(course_id);

CREATE INDEX IF NOT EXISTS course_students_access_paths_student_id_idx ON public.course_students_access_paths(student_id);

CREATE INDEX IF NOT EXISTS course_students_access_paths_location_id_idx ON course_students_access_paths(location_id);

CREATE INDEX IF NOT EXISTS study_plans_course_id_idx ON public.study_plans(course_id);

CREATE INDEX IF NOT EXISTS student_latest_submissions_student_id_idx ON student_latest_submissions(student_id);

ALTER TABLE public.student_packages DROP CONSTRAINT fk__student_packages__package_id;

ALTER TABLE public.student_packages ADD CONSTRAINT fk_student_packages_package_id FOREIGN KEY(package_id) REFERENCES package(package_id);

ALTER TABLE public.student_course DROP CONSTRAINT student_course_student_package_id_fk;

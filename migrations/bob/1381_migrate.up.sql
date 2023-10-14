ALTER TABLE public.package_course DROP CONSTRAINT IF EXISTS fk__package_course__course_id;

ALTER TABLE public.student_product DROP CONSTRAINT IF EXISTS fk__student_product__location_id;
ALTER TABLE public.student_product DROP CONSTRAINT IF EXISTS fk__student_product__product_id;
ALTER TABLE public.student_product DROP CONSTRAINT IF EXISTS fk__student_product__student_id;
ALTER TABLE public.student_product DROP CONSTRAINT IF EXISTS fk__student_product__updated_from_student_product_id;
ALTER TABLE public.student_product DROP CONSTRAINT IF EXISTS fk__student_product__updated_to_student_product_id;

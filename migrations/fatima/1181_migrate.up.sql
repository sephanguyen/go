DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'product_discount') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.product_discount;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'upcoming_student_course') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.upcoming_student_course;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'student_course') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.student_course;
  END IF;
END;
$$

DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'student_associated_product') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.student_associated_product;
  END IF;
END;
$$

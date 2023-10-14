DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'product_grade') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.product_grade;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'product_price') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.product_price;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'package_course_fee') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.package_course_fee;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'package_course_material') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.package_course_material;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'product_accounting_category') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.product_accounting_category;
  END IF;
END;
$$

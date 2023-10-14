DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'courses') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.courses;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'course_access_paths') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.course_access_paths;
  END IF;
END;
$$

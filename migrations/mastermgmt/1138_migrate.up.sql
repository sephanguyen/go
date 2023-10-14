DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'course_academic_year') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.course_academic_year;
  END IF;
END;
$$;

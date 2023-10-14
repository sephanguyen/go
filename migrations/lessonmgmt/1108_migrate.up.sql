DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='alloydb_publication') THEN
      CREATE PUBLICATION alloydb_publication;
   END IF;
END
$do$;

DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'lessons_teachers') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lessons_teachers;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'lessons_courses') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lessons_courses;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'reallocation') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.reallocation;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'lesson_reports') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lesson_reports;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'lesson_report_details') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lesson_report_details;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'partner_form_configs') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.partner_form_configs;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'partner_dynamic_form_field_values') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.partner_dynamic_form_field_values;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'classroom') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.classroom;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'lessons') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lessons;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'lesson_members') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.lesson_members;
  END IF;
END;
$$;

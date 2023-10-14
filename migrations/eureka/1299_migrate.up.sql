CREATE
OR REPLACE FUNCTION public.get_lo_list_in_course(course_id text) RETURNS SETOF learning_material LANGUAGE sql STABLE AS $function$
SELECT
  lm.*
FROM
  courses_books AS cb
  JOIN book_tree bt ON cb.book_id = bt.book_id
  JOIN learning_material lm ON bt.learning_material_id = lm.learning_material_id
WHERE
  cb.course_id = course_id $function$

CREATE INDEX IF NOT EXISTS
    courses__search_by_name__idx
    ON public.courses (
                       created_at DESC NULLS FIRST,
                       display_order ASC NULLS LAST,
                       left(name, 256) ASC NULLS LAST,
                       left(course_id, 256) ASC NULLS LAST
        );

DROP FUNCTION IF EXISTS search_courses_by_name;

CREATE OR REPLACE FUNCTION public.search_courses_by_name(search_name text, search_limit int, search_offset int) RETURNS SETOF public.courses
    LANGUAGE sql
    STABLE
AS
$$
SELECT *
FROM courses
WHERE (
          (courses.name) ILIKE ('%' || search_name || '%')
          )
ORDER BY created_at DESC NULLS FIRST,
         display_order ASC NULLS LAST,
         left(name, 256) ASC NULLS LAST,
         left(course_id, 256) ASC NULLS LAST
LIMIT search_limit OFFSET search_offset
$$;
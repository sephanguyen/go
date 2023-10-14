DROP FUNCTION IF EXISTS get_all_group_report_by_lesson_id;

CREATE OR REPLACE FUNCTION public.get_all_group_report_by_lesson_id(lesson_id_query text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lessons l on l.lesson_id = lr.lesson_id
where l.class_id = (
    select distinct class_id
      from lessons l1 where l1.lesson_id = lesson_id_query 
      and l1.deleted_at is NULL
      limit 1)
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
order by
    l.start_time desc
$$;

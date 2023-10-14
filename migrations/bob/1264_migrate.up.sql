DROP FUNCTION IF EXISTS get_previous_lesson_report_group;

CREATE OR REPLACE FUNCTION public.get_previous_lesson_report_group(lesson_id_query text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lessons l on l.lesson_id = lr.lesson_id
where l.start_time < (
    select l1.start_time 
      from lessons l1 where l1.lesson_id = lesson_id_query 
      and l1.class_id = l.class_id
      and l1.deleted_at is NULL limit 1)
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
order by
    l.start_time desc
limit 1;
$$;